package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	_ "github.com/go-sql-driver/mysql"
)

type CmdArgs struct {
	DumpServiceURL   string
	LookupServiceURL string
	Key              string
	DBUser           string
	DBName           string
	DBPassword       string
	UbikomName       string
	ReportRecipients []string
}

type Registration struct {
	Name      string
	Timestamp string
}

type ReportArgs struct {
	RegNum                   int
	Registrations            []*Registration
	IMAPClientNum            int
	POPClientNum             int
	ClientNum                int
	NewClientNum             int
	SMTPMessagesSent         int
	ExternalMessagesSent     int
	ExternalMessagesReceived int
	Fortune                  string
	DiskInfo                 string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	args := &CmdArgs{}
	flag.StringVar(&args.DumpServiceURL, "dump-service-url", globals.PublicDumpServiceURL, "dump service URL")
	flag.StringVar(&args.LookupServiceURL, "lookup-service-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.StringVar(&args.Key, "key", "", "key location")
	flag.StringVar(&args.DBUser, "db-user", "ubikom", "database user")
	flag.StringVar(&args.DBName, "db-name", "ubikom", "name of the database to open")
	flag.StringVar(&args.DBPassword, "db-password", "", "database password")
	flag.StringVar(&args.UbikomName, "ubikom-name", "", "ubikom name")
	var recipients string
	flag.StringVar(&recipients, "recipients", "", "report recipients")
	flag.Parse()

	if len(recipients) > 0 {
		args.ReportRecipients = strings.Split(recipients, ",")
	}

	privateKey, ubikomName, err := ReadKey(args.Key)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load private key")
	}

	// We take ubikom name from the key file name, unless it was explicitly specified.
	if args.UbikomName != "" {
		ubikomName = args.UbikomName
	}

	ctx := context.Background()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}
	lookupConn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()
	lookupService := pb.NewLookupServiceClient(lookupConn)

	db, err := OpenDB(args.DBUser, args.DBPassword, args.DBName)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open the database")
	}

	reportArgs := &ReportArgs{}
	reg, err := GetRegistrations(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get registration")
	}
	reportArgs.RegNum = len(reg)
	reportArgs.Registrations = reg

	reportArgs.IMAPClientNum, err = GetIMAPClientNum(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get imap clients num")
	}

	reportArgs.POPClientNum, err = GetPOPClientNum(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get pop clients num")
	}

	reportArgs.ClientNum, err = GetClientNum(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get pop clients num")
	}

	reportArgs.NewClientNum, err = GetNewClientNum(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get new clients num")
	}

	reportArgs.SMTPMessagesSent, err = GetSMTPMessagesSent(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get smtp messages sent")
	}

	reportArgs.ExternalMessagesSent, err = GetExternalMessagesSent(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get external messages sent")
	}

	reportArgs.ExternalMessagesReceived, err = GetExternalMessagesReceived(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get external messages received")
	}

	reportArgs.Fortune, err = GetFortune()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get fortune")
	}

	reportArgs.DiskInfo, err = GetDiskInfo()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get disk info")
	}

	report, err := GenerateReport(reportArgs)
	if err != nil {
		log.Fatal().Err(err).Msg("error generating report")
	}

	// If there are no recipients, just print out the report.
	if len(args.ReportRecipients) == 0 {
		fmt.Printf("%s\n", report)
		return
	}

	for _, r := range args.ReportRecipients {
		message := mail.NewMessage(r, ubikomName, "Ubikom Daily Report", report)
		err = protoutil.SendEmail(ctx, privateKey, []byte(message), ubikomName, r, lookupService)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send daily report")
		}
	}
}

func ReadKey(keyFile string) (*easyecc.PrivateKey, string, error) {
	var err error
	if keyFile == "" {
		keyFile, err = util.GetDefaultKeyLocation()
		if err != nil {
			return nil, "", err
		}
	}

	encrypted, err := util.IsKeyEncrypted(keyFile)
	if err != nil {
		return nil, "", err
	}

	passphrase := ""
	if encrypted {
		passphrase, err = util.ReadPassphase()
		if err != nil {
			return nil, "", err
		}
	}

	privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
	if err != nil {
		return nil, "", err
	}

	name := util.FileNameNoExtension(keyFile)
	return privateKey, name, nil
}

func OpenDB(user, password string, dbName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", user, password, dbName))
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, err
}

func GetRegistrations(db *sql.DB) ([]*Registration, error) {
	const query = `
SELECT 
	data1, timestamp 
FROM 
	events
WHERE
	event_type = 'ET_NAME_REGISTRATION' AND
	timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
ORDER BY
	timestamp
`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var res []*Registration
	for rows.Next() {
		reg := &Registration{}
		err = rows.Scan(&reg.Name, &reg.Timestamp)
		if err != nil {
			return nil, err
		}
		res = append(res, reg)
	}
	return res, nil
}

func GetPOPClientNum(db *sql.DB) (int, error) {
	const query = `
SELECT 
	COUNT(DISTINCT user1)
FROM 
	events
WHERE
	  event_type = 'ET_PROXY_POP_LOGIN' AND
	  timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GetIMAPClientNum(db *sql.DB) (int, error) {
	const query = `
SELECT 
	COUNT(DISTINCT user1)
FROM 
	events
WHERE
	  event_type = 'ET_PROXY_IMAP_LOGIN' AND
	  timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GetClientNum(db *sql.DB) (int, error) {
	const query = `
SELECT 
	COUNT(DISTINCT user1)
FROM 
	events
WHERE
	  (event_type = 'ET_PROXY_POP_LOGIN' OR 
	  	 event_type = 'ET_PROXY_IMAP_LOGIN') AND
	  timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GetNewClientNum(db *sql.DB) (int, error) {
	const query = `
SELECT
	COUNT(DISTINCT user1)
FROM
	events
WHERE	
	user1 NOT IN (
	        SELECT
			DISTINCT user1
	        FROM
			events
		WHERE
			(event_type = 'ET_PROXY_POP_LOGIN' or event_type = 'ET_PROXY_IMAP_LOGIN') AND
  			timestamp < DATE_ADD(NOW(), INTERVAL -1 DAY)
	) AND
		(event_type = 'ET_PROXY_POP_LOGIN' or event_type = 'ET_PROXY_IMAP_LOGIN') AND
		timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GetSMTPMessagesSent(db *sql.DB) (int, error) {
	const query = `
SELECT
	COUNT(*)
FROM
	events
WHERE
		event_type = 'ET_PROXY_SMTP_MESSAGE_SENT' AND
		timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GetExternalMessagesSent(db *sql.DB) (int, error) {
	const query = `
SELECT
	COUNT(*)
FROM
	events
WHERE
		event_type = 'ET_GATEWAY_EMAIL_MESSAGE_SENT' AND
		timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GetExternalMessagesReceived(db *sql.DB) (int, error) {
	const query = `
SELECT
	COUNT(*)
FROM
	events
WHERE
		event_type = 'ET_GATEWAY_UBIKOM_MESSAGE_SENT' AND
		timestamp BETWEEN DATE_ADD(NOW(), INTERVAL -1 DAY) AND NOW()
`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	var num int
	if !rows.Next() {
		return 0, fmt.Errorf("no data found")
	}
	err = rows.Scan(&num)
	if err != nil {
		return 0, nil
	}
	return num, nil
}

func GenerateReport(args *ReportArgs) (string, error) {
	const reportTmplTxt = `Greetings humans!

This is Ubikom Report Generator, broadcasting live from Ubikom world headquaters, and boy,
do I have some stats for you!

In the last 24 hours, we had {{.RegNum}} new registrations.

We also had {{.ClientNum}} clients actually using the service, which includes {{.IMAPClientNum}} IMAP clients
and {{.POPClientNum}} POP clients.

There were {{.NewClientNum}} new clients who actually used the service.

There were {{.SMTPMessagesSent}} messages sent via SMTP.

{{.ExternalMessagesSent}} messages were sent to external recipients.

{{.ExternalMessagesReceived}} messages were received from external users.

Here's the list of names that were registered:

{{range .Registrations}}
Name: {{printf "%-20s" .Name}} Time: {{.Timestamp}}
{{end}}

Until next time,

	Ubikom Report Generator

P.S. 

{{.Fortune}}

--- Disk Space ---

{{.DiskInfo}}
`
	reportTmpl, err := template.New("report").Parse(reportTmplTxt)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = reportTmpl.Execute(&b, args)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func GetFortune() (string, error) {
	cmd := exec.Command("bash", "-c", "fortune | cowsay")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func GetDiskInfo() (string, error) {
	cmd := exec.Command("/bin/df", "-h")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", err
	}
	return out.String(), nil
}

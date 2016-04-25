package mailer

import (
    "log"
    "time"
    "fmt"
    "net/smtp"
    "crypto/tls"
    "../utils"
    "io/ioutil"
    "encoding/base64"
)

type EmailUser struct {
    Username string
    EmailAddress string
    Password string
    RealName string
}

type ServerConfig struct {
    Port string
    Server string
    SkipCertVerify bool
}

type RcptList []string
type Filename string
type Attachment utils.Pair

type Email struct {
    SendText string
    Attachments []Attachment // base64 enc
    Rcpt RcptList
    Cc []string
    Delimiter string
    Subject string
}

func ( r *RcptList ) Set ( value string ) error {
    // we get some string and want to append it to r
    // TODO sanity check of mail
    *r = append(*r, value)
    return nil
}

func ( r *RcptList ) String () (ret string){
    for index, rcpt := range *r {
        if index != 0 {
            ret += ","
        }
        ret += " <" + rcpt + ">"
    }
    return
}

func SendMail (config ServerConfig, user EmailUser, mail Email) {
    // Set up authentication information.
    auth := smtp.PlainAuth(
        "",
        user.Username,
        user.Password,
        config.Server,
    )

    tlsconfig := &tls.Config {
        InsecureSkipVerify: true,
        ServerName: config.Server,
    }

    sconn, err := smtp.Dial(config.Server + ":" + config.Port)
    handleIfError(err)

    err = sconn.StartTLS(tlsconfig)
    handleIfError(err)

    err = sconn.Auth(auth)
    handleIfError(err)

    err = sconn.Mail(user.EmailAddress)
    handleIfError(err)

    for _, rcpt := range mail.Rcpt {
        err = sconn.Rcpt(rcpt)
        handleIfError(err)
    }

    writer, err := sconn.Data()
    handleIfError(err)

    log.Println(fmt.Sprintf("Sending E-Mail to %s including %d Attachments with Subject %s",
        mail.Rcpt.String(), len(mail.Attachments), mail.Subject))
    _, err = writer.Write([]byte(buildHeader(user, mail) + buildBody(mail) + buildAttachment(mail)))
    handleIfError(err)

    err = writer.Close()
    handleIfError(err)

    sconn.Quit()
}

func handleIfError (err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func buildHeader (from EmailUser, email Email) (ret string) {
    ret += fmt.Sprintf("From: %s <%s>\r\n", from.RealName, from.EmailAddress)
    ret += fmt.Sprintf("To: %s\r\n", email.Rcpt.String())
    ret += fmt.Sprintf("Subject: %s\r\n", email.Subject)
    ret += fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC822Z))
    ret += fmt.Sprintf("MIME-Version: 1.0\r\n")
    if len(email.Attachments) == 0 {
        ret += fmt.Sprintf("Content-Type: text/plain; charset=\"UTF-8\"")
    } else {
        ret += fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n--%s",
            email.Delimiter, email.Delimiter)
    }
    return
}

func buildBody (email Email) (ret string) {
    if len(email.Attachments) != 0 {
        ret += fmt.Sprintf("\r\nContent-Type: text/plain")
    }
    ret += fmt.Sprintf("\r\nContent-Transfer-Encoding:8bit\r\n\r\n")
    ret += fmt.Sprintf("%s\r\n", email.SendText,)
    if len(email.Attachments) != 0 {
        ret += fmt.Sprintf("--%s", email.Delimiter)
    }
    return
}

func buildAttachment (email Email) (ret string) {
    for index, pair := range email.Attachments {
        // read the files
        b, err := ioutil.ReadFile(pair.B)
        handleIfError(err)
        str := base64.StdEncoding.EncodeToString(b)
        ret += fmt.Sprintf("\r\nContent-Type: text/plain;")
        ret += fmt.Sprintf("name=\"%s\"\r\n", pair.A)
        ret += fmt.Sprintf("Content-Transfer-Encoding:base64\r\n")
        ret += fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", pair.A)
        ret += fmt.Sprintf("\r\n%s\r\n--%s", str, email.Delimiter)
        if index == len(email.Attachments) {
            ret += fmt.Sprintf("\r\n--%s--", email.Delimiter)
        }
    }
    return
}

func NewEmail () *Email {
    return &Email{ Delimiter : "AdhaueFDfAuq1243asd1" } // some random delimiter
}


package main

import (
    "io/ioutil"
    "fmt"
    "./mailer"
    "flag"
    "log"
    "./fileprocessing"
)

var flags CmdFlags

type CmdFlags struct {
    to mailer.RcptList
    message *string
    mailTemplateFile *string
    subject *string
    gitFolderPath *string
    folderNamingPrefix *string
    fileSelectorNamingSuffix *string
    recurseThroughSubDirs *bool
    keepFolderStructure *bool
    sendCopy *bool
    emailUser *string
    emailAddress *string
    emailServer *string
    emailPort *string
    emailPassword *string
    footNote *string
    skipCertVerify *bool
}

func (flags CmdFlags) String () string {
    msg := ""

    if flags.message != nil {
        msg = *(flags.message)
    }

    return fmt.Sprintf("\nmessage: %s, \nsendCopy: %t",
        msg, flags.sendCopy)
}

func main () {
    flags = CmdFlags{
        message : flag.String("message", "", "message to send if no mailTemplateFile is specified"),
        mailTemplateFile : flag.String("template", "", "path to some text file to be used as tempalte"),
        subject : flag.String("subject", "", "email subject"),
        gitFolderPath : flag.String("gitFolderPath", "", "Path to folder to be processed"),
        folderNamingPrefix : flag.String("folderNamingPrefix", "",
            `Prefix to be used when selecting folders. The highest number
            in de suffix will be used as the folder to recurse`),
        recurseThroughSubDirs : flag.Bool("recurseThroughSubDirs", false, "look through sub dirs"),
        sendCopy : flag.Bool("sendCopy", true, "send copy to other to parameters"),
        fileSelectorNamingSuffix : flag.String("fileSelectorNamingSuffix", "", "file suffix to be selected."),
        emailUser : flag.String("emailUser", "", "email username"),
        emailAddress : flag.String("emailAddress", "", "real from email address"),
        emailServer : flag.String("emailServer" ,"", ""),
        emailPort : flag.String("emailPort", "", ""),
        emailPassword : flag.String("emailPassword", "", ""),
        skipCertVerify : flag.Bool("SkipCertVerify", true, "skip tls verification"),
        footNote : flag.String("footNote", getDefaultFootnote(), "footnote to be send in email"),
    }
    flag.Var(&flags.to, "to", "one mail address of rcpt. might be used multiple times")

    flag.Parse()
    handleFlags()
    email := buildEmail()
    config := mailer.ServerConfig{
        Server : *flags.emailServer,
        Port : *flags.emailPort,
        SkipCertVerify : *flags.skipCertVerify,
    }
    user := mailer.EmailUser{
        Username : *flags.emailUser,
        EmailAddress : *flags.emailAddress,
        Password : *flags.emailPassword,
    }

    mailer.SendMail(config, user, *email)
}

func getDefaultFootnote () string {
    return "\r\n\r\n--\r\n\r\n" +
           "This email was automatically generated\r\n" +
           "If you received this mail in error please contact \r\n" +
           "someone else who is addressed by this email. We \r\n" +
           "will take care that this will not happen again."
}

func buildEmail () (email *mailer.Email) {
    email = mailer.NewEmail()
    email.Rcpt = make([]string, len(flags.to))
    for index, parsed := range flags.to {
        email.Rcpt[index] = parsed
    }
    if *flags.mailTemplateFile == "" {
        email.SendText = *flags.message
    } else {
        b, err := ioutil.ReadFile(*flags.mailTemplateFile)
        if err != nil {
            handleError("Incorrect template file path.")
        }
        email.SendText = string(b)
    }
    email.SendText += *flags.footNote
    attachments, version := fileprocessing.GetAttachements( *flags.folderNamingPrefix,
        *flags.fileSelectorNamingSuffix, *flags.gitFolderPath, *flags.recurseThroughSubDirs)
    email.Attachments = attachments
    email.Subject = fmt.Sprintf("%s %d", *flags.subject, version)
    return
}

func handleFlags () {
    if *flags.emailAddress == "" || *flags.emailServer == "" || *flags.emailPort == "" || *flags.emailPassword == "" || *flags.emailUser == "" {
        handleError("Email Server connection not fully specified.")
    }
    if flags.to == nil {
        handleError("no rcpt.")
    }
}

func handleError (err string) {
    flag.PrintDefaults()
    log.Fatal(err)
}


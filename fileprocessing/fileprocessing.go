package fileprocessing

import (
    "../mailer"
    "os"
    "strings"
    "strconv"
    "log"
)

func GetAttachements ( folderPrefix, fileSuffix, directory string, recurse bool ) ([]mailer.Attachment, int) {
    // take all folders matching folderPrefix*
    // sort by rest of foldername
    // take folder with highest version number
    // take all files matching fileSuffix in there
    // recurse there if necessary
    f, _ := os.Open(directory)
    infos, _ := f.Readdir(0)
    var maxVersion int = -1
    var maxInfo os.FileInfo
    for _, info := range infos {
        if info.IsDir() && strings.HasPrefix(info.Name(), folderPrefix) {
            version, err := strconv.Atoi(strings.TrimPrefix(info.Name(), folderPrefix))
            if err != nil {
                log.Fatal("Bad versioning of folders, expected numbers")
            }
            if version > maxVersion {
                maxVersion = version
                maxInfo = info
            }
        }
    }
    f.Close()
    if maxVersion == -1 {
        log.Fatal("Could not find any folders to match against given prefix")
    }
    searchDir := directory + "/" + maxInfo.Name()
    log.Println("Visiting folder " + searchDir + " searching for " + fileSuffix)
    return createAttachments(fileSuffix, searchDir, recurse), maxVersion
}

func createAttachments (fileSuffix, directory string, recurse bool) (att []mailer.Attachment) {
    f, _ := os.Open(directory)
    infos, _ := f.Readdir(0) // read all fileinfo in directory
    for _, finfo := range infos {
        log.Println("Inspecting descriptor " + finfo.Name())
        if finfo.IsDir() && recurse {
            createAttachments(fileSuffix, directory + "/" + finfo.Name(), recurse)
        } else {
            // it's a file. match it against suffix
            if strings.HasSuffix(finfo.Name(), fileSuffix) {
                // add to attachments
                att = append(att, mailer.Attachment{A : finfo.Name(), B : directory + "/" + finfo.Name()})
            }
        }
    }
    return
}

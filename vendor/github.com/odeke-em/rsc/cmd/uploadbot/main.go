// Uploadbot uploads tgz snapshots of Mercurial repositories to the download
// section of a Google Code project.
//
// Usage
//
// Synopsis:
//
//	uploadbot [-f] [-pw=pwfile] [-root=rootdir] [project...]
//
// Uploadbot reads from pwfile (default $HOME/codebot.pw) an email address
// and code.google.com-generated password in JSON format:
//
//	{"User": "bot@gmail.com", "Password": "3uiarglaer4rq"}
//
// It then uploads each of the named projects, which should already be checked
// out into subdirectories of rootdir (default $HOME/googlecode.upload) named
// for the projects. For example, code.google.com/p/re2 should be checked out
// into rootdir/re2.
//
// If no projects are given on the command line, uploadbot behaves as if all the
// subdirectories in rootdir were given.
//
// Uploadbot assumes that the checked-out directory for a project corresponds
// to the most recent upload. If there are no new changes to incorporate, as reported
// by "hg incoming", then uploadbot will not upload a new snapshot. The -f flag
// overrides this, forcing uploadbot to upload a new snapshot.
//
// The uploaded snapshot files are named project-yyyymmdd.tgz.
//
// Initial Setup
//
// First, find your generated password at https://code.google.com/hosting/settings
// and create $HOME/codebot.pw (chmod 600) in the form given above.
//
// Next, create the work directory for the upload bot:
//
//	mkdir $HOME/googlecode.upload
//
// Adding A Project
//
// To add a project, first check out the repository in the work directory:
//
//	cd $HOME/googlecode.upload
//	hg clone https://code.google.com/p/yourproject
//
// Then force the initial upload:
//
//	uploadbot -f yourproject
//
// Cron
//
// A nightly cron entry to upload all projects that need uploading at 5am would be:
//
//	0 5 * * *        /home/you/bin/uploadbot
//
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

var (
	pw    = flag.String("pw", os.Getenv("HOME")+"/codebot.pw", "file containing User/Password json")
	root  = flag.String("root", os.Getenv("HOME")+"/googlecode.upload", "directory of checked-out google code projects")
	force = flag.Bool("f", false, "force upload, even if nothing has changed")
)

var bot struct {
	User     string
	Password string
}

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*pw)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(data, &bot); err != nil {
		log.Fatalf("reading %s: %v", *pw, err)
	}

	dirs := flag.Args()
	if len(dirs) == 0 {
		all, err := ioutil.ReadDir(*root)
		if err != nil {
			log.Fatal(err)
		}
		for _, fi := range all {
			if fi.IsDir() {
				dirs = append(dirs, fi.Name())
			}
		}
	}

	for _, dir := range dirs {
		dir := path.Join(*root, dir)
		cmd := exec.Command("hg", "incoming")
		cmd.Dir = dir
		_, err := cmd.CombinedOutput()
		if err != nil && !*force {
			// non-zero means nothing incoming
			continue
		}

		fmt.Fprintf(os.Stderr, "uploading %s\n", dir)
		cmd = exec.Command("hg", "pull", "-u")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "sync %s: %v\n%s\n", dir, err, out)
			continue
		}

		f, err := ioutil.TempFile("", "uploadbot")
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating temp file: %v\n", err)
			continue
		}

		cmd = exec.Command("tar", "czf", f.Name(), path.Base(dir))
		cmd.Dir = path.Dir(dir)
		out, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tar %s: %v\n%s\n", dir, err, out)
			continue
		}

		err = upload(path.Base(dir), f)
		os.Remove(f.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "upload %s: %s\n", dir, err)
			continue
		}
	}
}

func upload(project string, f *os.File) error {
	now := time.Now()
	filename := fmt.Sprintf("%s-%s.tgz", project, now.Format("20060102"))
	summary := now.Format("source tree as of 2006-01-02")

	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	if err := w.WriteField("summary", summary); err != nil {
		return err
	}
	fw, err := w.CreateFormFile("filename", filename)
	if err != nil {
		return err
	}
	f.Seek(0, 0)
	if _, err = io.Copy(fw, f); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	// Send the file to Google Code.
	url := fmt.Sprintf("https://%s.googlecode.com/files", project)
	println(url)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	token := fmt.Sprintf("%s:%s", bot.User, bot.Password)
	token = base64.StdEncoding.EncodeToString([]byte(token))
	req.Header.Set("Authorization", "Basic "+token)
	req.Header.Set("Content-type", w.FormDataContentType())

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		fmt.Fprintf(os.Stderr, "%s upload failed:\n", project)
		io.Copy(os.Stderr, resp.Body)
		return fmt.Errorf("upload: %s", resp.Status)
	}
	return nil
}

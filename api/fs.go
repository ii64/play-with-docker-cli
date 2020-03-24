package api

/* Endpoints
FileSystem Tree
> GET https://labs.play-with-docker.com/sessions/bprhvsfnctv000ftgi1g/instances/bprhvsfn_bprip5fnctv000el7hgg/fstree

FileSystem Get Content
> GET https://labs.play-with-docker.com/sessions/bprhvsfnctv000ftgi1g/instances/bprhvsfn_bprip5fnctv000el7hgg/file?path=/root/authorized_keys

FileSystem Upload
> POST https://labs.play-with-docker.com/sessions/bprhvsfnctv000ftgi1g/instances/bprhvsfn_bprip5fnctv000el7hgg/uploads?path=/root

FileSystem Exec
> POST https://labs.play-with-docker.com/sessions/bprhvsfnctv000ftgi1g/instances/bprhvsfn_bprip5fnctv000el7hgg/exec

FileSystem Remove
> not implemented
*/
import (
	"fmt"
	"encoding/json"
	"encoding/base64"
	"net/http"
	"strings"

	"bytes"
	"mime/multipart"
	"path/filepath"
	"io"
)

var (
	ENDPOINT_FSTREE  = "/sessions/%s/instances/%s/fstree"
	ENDPOINT_EXEC    = "/sessions/%s/instances/%s/exec"
	ENDPOINT_UPLOAD  = "/sessions/%s/instances/%s/uploads?path=%s" //path=/dst/dir
	ENDPOINT_CONTENT = "/sessions/%s/instances/%s/file?path=%s"//path=/to/file
)

type fsItem struct {
	Type     string    `json:"type,omitempty"`
	Name     string    `json:"name,omitempty"`
	Contents []fsItem  `json:"contents,omitempty"`
	Target   string    `json:"target,omitempty"`
}
func (d *dashboard) FSCat(instanceId string, path string) ([]byte, error) {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_CONTENT, d.url.Host, d.session_id, instanceId, path)
	d.Log("url: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s or %s", ErrSessionNotFound, ErrInstanceNotFound)		
	}else if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrInternalError
	}else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}
	decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
	buf := make([]byte, 2)
	data := []byte{}
	for {
		n, err := decoder.Read(buf)
		if n == 0 || err != nil {
			break
		}
		data = append(data, buf[:n]...)
	}
	return data, nil
}

type FSTreeResponse []fsItem
func (s FSTreeResponse) GetDir(path string) (*fsItem, error) {
	splitted := strings.Split(path, "/")
	filtered := []string{}
	for _, i := range splitted {
		if len(i) != 0 {
			i = strings.Replace(i, "/", "", -1)
			i = strings.Replace(i, "\\", "", -1)
			filtered = append(filtered, i)
		}
	}
	var targ []fsItem = s
	var lastFd2 *fsItem
	var lastFd *fsItem
	for _, dirName := range filtered {
		for _, fd := range targ {
			if fd.Type == "directory" {
				if strings.Contains(fd.Name, dirName) {
					targ = fd.Contents
					lastFd = &fd
					break
				}else{
					continue
				}
			}else{
				continue
			}
		}
		if lastFd == nil || (lastFd2 != nil && lastFd2.Name == lastFd.Name) {
			return nil, fmt.Errorf("not found dir /%s", strings.Join(filtered, "/"))
		}
		lastFd2 = lastFd
	}
	lastFd.Name = fmt.Sprintf("/%s", strings.Join(filtered, "/"))
	return lastFd, nil

}
func (s FSTreeResponse) handleFile(r fsItem, compose *string, space *int) {
	*compose += strings.Repeat(" ", *space)
	*compose += r.Name
	*compose += "\n"
}
func (s FSTreeResponse) handleLink(r fsItem, compose *string, space *int) {
	*compose += strings.Repeat(" ", *space)
	*compose += r.Name
	*compose += " (link)\n"
}
func (s FSTreeResponse) handleDir(r fsItem, compose *string, space *int) {
	*compose += strings.Repeat(" ", *space)
	*compose += r.Name
	if r.Name[len(r.Name)-1] != 47 {
		*compose += "/"
	}
	*compose += "\n"
	*space += 3
	if len(r.Contents) == 0 {
		*compose += strings.Repeat(" ", *space)
		*compose += "(empty dir)\n"
	}
	for _, fd := range r.Contents {
		if fd.Type == "directory" {
			s.handleDir(fd, compose, space)
			*space -= 3
		}else if fd.Type == "file" {
			s.handleFile(fd, compose, space)
		}
	}
}
func (s FSTreeResponse) ToString(instanceId string) string{
	compose := instanceId + ":\n"
	space := 0
	for _, fi := range s {
		if fi.Type == "directory" {
			s.handleDir(fi, &compose, &space)
		}else if fi.Type == "file" {
			s.handleFile(fi, &compose, &space)
		}else if fi.Type == "link" {
			s.handleLink(fi, &compose, &space)
		}
	}
	return compose
}
func (d *dashboard) FSTree(instanceId string) (*FSTreeResponse, error) {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_FSTREE, d.url.Host, d.session_id, instanceId)
	d.Log("url: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s or %s", ErrSessionNotFound, ErrInstanceNotFound)
	}else if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrInternalError
	}else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}

	dx := &FSTreeResponse{}
	err = json.NewDecoder(resp.Body).Decode(dx)
	if err != nil {
		return nil, err
	}
	return dx, nil
}
type execRequest struct {
	Cmd []string `json:"command,omitempty"`
}
type execRespose struct {
	ExitCode int `json:"status_code,omitempty"`
}
func (d *dashboard) Exec(instanceId string, cmd []string) (*execRespose, error) {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_EXEC, d.url.Host, d.session_id, instanceId)
	d.Log("url: %s\n", url)
	payload := &execRequest{
		Cmd: cmd,
	}
	r, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s or %s", ErrSessionNotFound, ErrInstanceNotFound)
	}else if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrInternalError
	}else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}

	i := &execRespose{}
	err = json.NewDecoder(resp.Body).Decode(i)
	return i, nil
}
func (d *dashboard) FSPut(instanceId, dst string, content []byte) error {
	dirPath, fileName := filepath.Split(dst)
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_UPLOAD, d.url.Host, d.session_id, instanceId, dirPath)
	d.Log("url: %s\n", url)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("/"+fileName, fileName)
	if err != nil {
		return err
	}
	rd := bytes.NewReader(content)
	_, err = io.Copy(part, rd)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s or %s", ErrSessionNotFound, ErrInstanceNotFound)
	}else if resp.StatusCode == http.StatusInternalServerError {
		return ErrInternalError
	}else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}
	return nil
}
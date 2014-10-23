package netdisk

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
)

func WriteFile(token, filename string, body []byte) error {
	v := url.Values{}
	v.Add("method", "upload")
	v.Add("access_token", token)
	v.Add("path", filename)
	v.Add("ondup", "overwrite")

	raw := &bytes.Buffer{}
	writer := multipart.NewWriter(raw)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Disposition",
		fmt.Sprintf("form-data; name=\"file\"; filename=\"%s\"", path.Base(filename)))

	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}

	_, err = part.Write(body)
	if err != nil {
		return err
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	targetUrl := "https://c.pcs.baidu.com/rest/2.0/pcs/file?" + v.Encode()

	resp, err := http.Post(targetUrl, contentType, raw)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Invalid respone status")
	}

	return nil
}
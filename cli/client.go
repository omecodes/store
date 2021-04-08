package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/store/auth"
)

func putAccess(clientApp *auth.ClientApp) error {
	endpoint := fmt.Sprintf("%s/auth/accesses", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(clientApp)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	return nil
}

func getAccesses(outputFilename string) error {
	endpoint := fmt.Sprintf("%s/auth/accesses", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	if outputFilename == "" {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		outputFilename = fmt.Sprintf("%s.accesses.json", u.Host)
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	done := false
	buf := make([]byte, 1024)

	file, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	for !done {
		n, err := rsp.Body.Read(buf)
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteAccess(accessID string) error {
	endpoint := fmt.Sprintf("%s/auth/access/%s", fullAPILocation(), accessID)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	return nil
}

func putUser(user *auth.UserCredentials) error {
	endpoint := fmt.Sprintf("%s/auth/users", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	encoded, err := json.Marshal(user)
	if err != nil {
		return err
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	return nil
}

func putCollections(collection *objects.Collection) error {
	endpoint := fmt.Sprintf("%s/objects/collections", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	encoded, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	return nil
}

func listCollections(outputFilename string) error {
	endpoint := fmt.Sprintf("%s/objects/collections", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	if outputFilename == "" {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		outputFilename = fmt.Sprintf("%s.objects-collections.json", u.Host)
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	done := false
	buf := make([]byte, 1024)

	file, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	for !done {
		n, err := rsp.Body.Read(buf)
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}

func putFileSource(source *files.Source) error {
	endpoint := fmt.Sprintf("%s/files/sources", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	encoded, err := json.Marshal(source)
	if err != nil {
		return err
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	return nil
}

func listFileSources(outputFilename string) error {
	endpoint := fmt.Sprintf("%s/files/sources", fullAPILocation())
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	if outputFilename == "" {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		outputFilename = fmt.Sprintf("%s.files-sources.json", u.Host)
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	done := false
	buf := make([]byte, 1024)

	file, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	for !done {
		n, err := rsp.Body.Read(buf)
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteFileSources(sourceID string) error {
	endpoint := fmt.Sprintf("%s/files/sources/%s", fullAPILocation(), sourceID)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	username, password, err := promptAuthentication()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != 200 {
		return errors.BadRequest(rsp.Status)
	}

	return nil
}

package admin

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

func putAccess(adminPassword string, access *auth.APIAccess) error {
	endpoint := fmt.Sprintf("%s/auth/access", server)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	encoded, err := json.Marshal(access)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
	}

	return nil
}

func getAccesses(adminPassword string, outputFilename string) error {
	endpoint := fmt.Sprintf("%s/auth/accesses", server)
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

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
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

func deleteAccess(adminPassword string, accessID string) error {
	endpoint := fmt.Sprintf("%s/auth/access/%s", server, accessID)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
	}

	return nil
}

func putCollections(adminPassword string, collection *objects.Collection) error {
	endpoint := fmt.Sprintf("%s/objects/collections", server)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	encoded, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
	}

	return nil
}

func listCollections(adminPassword string, outputFilename string) error {
	endpoint := fmt.Sprintf("%s/objects/collections", server)
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

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
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

func putFileSource(adminPassword string, source *files.Source) error {
	endpoint := fmt.Sprintf("%s/files/sources", server)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	encoded, err := json.Marshal(source)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
	}

	return nil
}

func listFileSources(adminPassword string, outputFilename string) error {
	endpoint := fmt.Sprintf("%s/files/sources", server)
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

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
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

func deleteFileSources(adminPassword string, sourceID string) error {
	endpoint := fmt.Sprintf("%s/files/sources/%s", server, sourceID)
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", adminPassword)
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
		return errors.New(rsp.Status)
	}

	return nil
}

package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

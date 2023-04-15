/*
	Copyright 2023 Loophole Labs

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		   http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/loopholelabs/cloudflare/pkg/bindings"
	"github.com/loopholelabs/cloudflare/pkg/models"
	"github.com/rs/zerolog"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
)

type Options struct {
	LogName string
	UserID  string
	Token   string
	Prefix  string
}

type Client struct {
	logger              *zerolog.Logger
	options             *Options
	workerURL           *url.URL
	authorizationHeader string
	context             context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
}

func New(options *Options, logger *zerolog.Logger) (*Client, error) {
	l := logger.With().Str(options.LogName, "Cloudflare").Logger()

	workerURL, err := url.Parse("https://api.cloudflare.com/client/v4/accounts/" + options.UserID + "/workers/scripts")
	if err != nil {
		return nil, err
	}

	authorizationHeader := fmt.Sprintf("Bearer %s", options.Token)

	ctx, cancel := context.WithCancel(context.Background())

	e := &Client{
		logger:              &l,
		options:             options,
		workerURL:           workerURL,
		authorizationHeader: authorizationHeader,
		context:             ctx,
		cancel:              cancel,
	}

	return e, nil
}

func (c *Client) Close() error {
	c.logger.Debug().Msg("closing cloudflare client")
	c.cancel()
	defer c.wg.Wait()
	return nil
}

func (c *Client) UploadScaleFunction(identifier string, wrapperScript []byte, functions []*bindings.ScaleFunction) (*models.UploadResponse, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	wrapperScriptReader := bytes.NewReader(wrapperScript)
	err := addPart(writer, "worker.js", "worker.js", "application/javascript", wrapperScriptReader)
	if err != nil {
		return nil, fmt.Errorf("error adding wrapper script to multipart request: %w", err)
	}

	for _, function := range functions {
		sfReader := bytes.NewReader(function.ScaleFunction.Encode())
		name := fmt.Sprintf("%s.bin", function.Identifier)
		err = addPart(writer, name, name, "application/octet-stream", sfReader)
		if err != nil {
			return nil, fmt.Errorf("error adding scale function to multipart request: %w", err)
		}

		wasmReader := bytes.NewReader(function.WASM)
		name = fmt.Sprintf("%s.wasm", function.Identifier)
		err = addPart(writer, name, name, "application/wasm", wasmReader)
		if err != nil {
			return nil, fmt.Errorf("error adding wasm module to multipart request: %w", err)
		}
	}

	workers := make([]bindings.Worker, 0, len(functions)*2)
	for _, function := range functions {
		workers = append(workers, bindings.Worker{
			Type: "data_blob",
			Name: fmt.Sprintf("__SF_%s", function.Identifier),
			Part: fmt.Sprintf("%s.bin", function.Identifier),
		})
		workers = append(workers, bindings.Worker{
			Type: "wasm_module",
			Name: fmt.Sprintf("__WASM_%s", function.Identifier),
			Part: fmt.Sprintf("%s.wasm", function.Identifier),
		})
	}

	metadata := bindings.Metadata{
		BodyPart: "worker.js",
		Bindings: workers,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("error marshaling metadata: %w", err)
	}
	err = addPart(writer, "metadata", "metadata.json", "application/json", bytes.NewReader(metadataJSON))
	if err != nil {
		return nil, fmt.Errorf("error adding metadata to multipart request: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %w", err)
	}

	requestURL := c.workerURL.String() + "/" + c.options.Prefix + identifier + "?include_subdomain_availability=true&excludeScript=true"
	req, err := http.NewRequest("PUT", requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating upload request: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", c.authorizationHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error uploading worker: %w", err)
	}
	if resp.StatusCode != 200 {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error uploading worker (%d: %s): %w", resp.StatusCode, resp.Status, err)
		}
		return nil, fmt.Errorf("error uploading worker (%d: %s): %s", resp.StatusCode, resp.Status, errBody)
	}

	res := new(models.UploadResponse)
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("error decoding upload response: %w", err)
	}
	return res, nil
}

func (c *Client) DeleteScaleFunction(identifier string) error {
	requestURL := c.workerURL.String() + "/" + c.options.Prefix + identifier
	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}
	req.Header.Add("Authorization", c.authorizationHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting worker: %w", err)
	}
	if resp.StatusCode != 200 {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error deleting worker (%d: %s): %w", resp.StatusCode, resp.Status, err)
		}
		return fmt.Errorf("error deleting worker (%d: %s): %s", resp.StatusCode, resp.Status, errBody)
	}
	return nil
}

func addPart(w *multipart.Writer, name string, filename string, contentType string, r io.Reader) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filename))
	h.Set("Content-Type", contentType)
	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, r)
	return err
}

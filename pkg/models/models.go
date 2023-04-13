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

package models

type UploadResponse struct {
	Success  bool            `json:"success"`
	Errors   []ResponseError `json:"errors"`
	Messages []ResponseError `json:"messages"`
	Result   ResponseResult  `json:"result"`
}

type ResponseResult struct {
	Id                   string   `json:"id"`
	CreatedOn            string   `json:"created_on"`
	ModifiedOn           string   `json:"modified_on"`
	Etag                 string   `json:"etag"`
	UsageModel           string   `json:"usage_model"`
	Handlers             []string `json:"handlers"`
	AvailableOnSubdomain bool     `json:"available_on_subdomain"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

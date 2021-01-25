// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by go-swagger; DO NOT EDIT.

package experiment_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	experiment_model "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
)

// CreateExperimentReader is a Reader for the CreateExperiment structure.
type CreateExperimentReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateExperimentReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewCreateExperimentOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewCreateExperimentDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCreateExperimentOK creates a CreateExperimentOK with default headers values
func NewCreateExperimentOK() *CreateExperimentOK {
	return &CreateExperimentOK{}
}

/*CreateExperimentOK handles this case with default header values.

A successful response.
*/
type CreateExperimentOK struct {
	Payload *experiment_model.APIExperiment
}

func (o *CreateExperimentOK) Error() string {
	return fmt.Sprintf("[POST /apis/v1beta1/experiments][%d] createExperimentOK  %+v", 200, o.Payload)
}

func (o *CreateExperimentOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(experiment_model.APIExperiment)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateExperimentDefault creates a CreateExperimentDefault with default headers values
func NewCreateExperimentDefault(code int) *CreateExperimentDefault {
	return &CreateExperimentDefault{
		_statusCode: code,
	}
}

/*CreateExperimentDefault handles this case with default header values.

CreateExperimentDefault create experiment default
*/
type CreateExperimentDefault struct {
	_statusCode int

	Payload *experiment_model.APIStatus
}

// Code gets the status code for the create experiment default response
func (o *CreateExperimentDefault) Code() int {
	return o._statusCode
}

func (o *CreateExperimentDefault) Error() string {
	return fmt.Sprintf("[POST /apis/v1beta1/experiments][%d] CreateExperiment default  %+v", o._statusCode, o.Payload)
}

func (o *CreateExperimentDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(experiment_model.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/cacheimpls"
	"iam/pkg/config"
	svctypes "iam/pkg/service/types"
)

const superSystemID = "SUPER"

// AnyExpression ...
var AnyExpression = map[string]interface{}{
	"op":    "any",
	"field": "",
	"value": []interface{}{},
}

func copyRequestFromAuthBody(req *request.Request, body *authRequest) {
	req.System = body.System

	req.Action.ID = body.Action.ID

	req.Subject.Type = body.Subject.Type
	req.Subject.ID = body.Subject.ID

	for _, resource := range body.Resources {
		req.Resources = append(req.Resources, types.Resource{
			System:    resource.System,
			Type:      resource.Type,
			ID:        resource.ID,
			Attribute: resource.Attribute,
		})
	}
}

func copyRequestFromQueryBody(req *request.Request, body *queryRequest) {
	req.System = body.System

	req.Action.ID = body.Action.ID

	req.Subject.Type = body.Subject.Type
	req.Subject.ID = body.Subject.ID

	for _, resource := range body.Resources {
		req.Resources = append(req.Resources, types.Resource{
			System:    resource.System,
			Type:      resource.Type,
			ID:        resource.ID,
			Attribute: resource.Attribute,
		})
	}
}

func copyRequestFromQueryByActionsBody(req *request.Request, body *queryByActionsRequest) {
	req.System = body.System

	req.Subject.Type = body.Subject.Type
	req.Subject.ID = body.Subject.ID

	for _, resource := range body.Resources {
		req.Resources = append(req.Resources, types.Resource{
			System:    resource.System,
			Type:      resource.Type,
			ID:        resource.ID,
			Attribute: resource.Attribute,
		})
	}
}

func copyRequestFromAuthByActionsBody(req *request.Request, body *authByActionsRequest) {
	req.System = body.System

	req.Subject.Type = body.Subject.Type
	req.Subject.ID = body.Subject.ID

	for _, resource := range body.Resources {
		req.Resources = append(req.Resources, types.Resource{
			System:    resource.System,
			Type:      resource.Type,
			ID:        resource.ID,
			Attribute: resource.Attribute,
		})
	}
}

func copyRequestFromAuthByResourcesBody(req *request.Request, body *authByResourcesRequest) {
	req.System = body.System

	req.Action.ID = body.Action.ID

	req.Subject.Type = body.Subject.Type
	req.Subject.ID = body.Subject.ID
}

func hasSystemSuperPermission(systemID, _type, id string) (bool, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "validateSystemSuperUser")

	// check default superuser
	if _type == svctypes.UserType && config.SuperUserSet.Has(id) {
		return true, nil
	}

	// check system manager or super manager
	systemIDs, err := cacheimpls.ListSubjectRoleSystemID(_type, id)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.ListSubjectRoleSystemID subjectType=`%s`, subjectID=`%s` fail",
			systemID, _type, id)
		return false, err
	}

	if len(systemIDs) == 0 {
		return false, nil
	}

	for _, s := range systemIDs {
		if s == systemID || s == superSystemID {
			return true, nil
		}
	}

	return false, nil
}

func buildResourceID(rs []resource) string {
	// single:  system,type,id
	// multiple: system,type,id/system,type,id
	if len(rs) == 0 {
		return ""
	}

	nodes := make([]string, 0, len(rs))
	for _, r := range rs {
		nodes = append(nodes, fmt.Sprintf("%s,%s,%s", r.System, r.Type, r.ID))
	}

	return strings.Join(nodes, "/")
}

// ValidateSystemMatchClient ...
func ValidateSystemMatchClient(systemID, clientID string) error {
	if systemID == "" || clientID == "" {
		return fmt.Errorf("system_id or client_id do not allow empty")
	}

	validClients, err := cacheimpls.GetSystemClients(systemID)
	if err != nil {
		return fmt.Errorf("get system(%s) valid clients fail, err=%w", systemID, err)
	}

	for _, c := range validClients {
		if clientID == c {
			return nil
		}
	}

	return fmt.Errorf("client(%s) can not request system(%s)", clientID, systemID)
}

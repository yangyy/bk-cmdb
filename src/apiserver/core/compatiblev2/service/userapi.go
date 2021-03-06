/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"strconv"
	"strings"

	"configcenter/src/apiserver/core/compatiblev2/common/converter"
	"configcenter/src/common"
	"configcenter/src/common/blog"
	ccError "configcenter/src/common/errors"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/common/util"

	"github.com/emicklei/go-restful"
)

func (s *service) getCustomerGroupList(req *restful.Request, resp *restful.Response) {

	srvData := s.newSrvComm(req.Request.Header)
	defErr := srvData.ccErr

	err := req.Request.ParseForm()
	if err != nil {
		blog.Errorf("getCustomerGroupList error:%v,rid:%s", err, srvData.rid)
		converter.RespFailV2(common.CCErrCommPostInputParseError, defErr.Error(common.CCErrCommPostInputParseError).Error(), resp)
		return
	}

	formData := req.Request.Form
	strAppIDs := formData.Get("ApplicationIDs")

	if "" == strAppIDs {
		blog.Error("getCustomerGroupList error: param ApplicationIDs is empty!input:%#v,rid:%s", formData, srvData.rid)
		converter.RespFailV2(common.CCErrCommParamsNeedSet, defErr.Errorf(common.CCErrCommParamsNeedSet, "ApplicationIDs").Error(), resp)
		return
	}

	appIDs := strings.Split(strAppIDs, ",")

	var postInput metadata.QueryInput
	postInput.Start = 0
	postInput.Limit = common.BKNoLimit

	resDataV2 := mapstr.NewArray()

	// all application ids
	for _, appID := range appIDs {

		result, err := s.CoreAPI.HostServer().GetUserCustomQuery(srvData.ctx, appID, srvData.header, &postInput)
		if err != nil {
			blog.Errorf("getCustomerGroupList http do error.err:%v,input:%#v,rid:%s", err, formData, srvData.rid)
			converter.RespFailV2(common.CCErrCommHTTPDoRequestFailed, defErr.Error(common.CCErrCommHTTPDoRequestFailed).Error(), resp)
			return
		}
		if !result.Result {
			blog.Errorf("getCustomerGroupList http reply error.reply:%#v,input:%#v,rid:%s", result, formData, srvData.rid)
			converter.RespFailV2(result.Code, result.ErrMsg, resp)
			return
		}

		//translate cmdb v3 to v2 api result
		retItem, err := converter.ResToV2ForCustomerGroup(result.Result, result.ErrMsg, result.Data, appID)

		//translate cmdb v3 to v2 api result error,
		if err != nil {
			blog.Errorf("getCustomerGroupList error:%s, reply:%v,input:%#v,rid:%s", err.Error(), result.Data, formData, srvData.rid)
			converter.RespFailV2(common.CCErrCommReplyDataFormatError, defErr.Error(common.CCErrCommReplyDataFormatError).Error(), resp)
			return
		}
		if 0 == len(retItem) {
			continue
		}
		resDataV2 = append(resDataV2, mapstr.MapStr{"ApplicationID": appID, "CustomerGroup": retItem})

	}

	converter.RespSuccessV2(resDataV2, resp)
}

func (s *service) getContentByCustomerGroupID(req *restful.Request, resp *restful.Response) {

	srvData := s.newSrvComm(req.Request.Header)
	defErr := srvData.ccErr

	err := req.Request.ParseForm()
	if err != nil {
		blog.Errorf("getContentByCustomerGroupID error:%v,rid:%s", err, srvData.rid)
		converter.RespFailV2(common.CCErrCommPostInputParseError, defErr.Error(common.CCErrCommPostInputParseError).Error(), resp)
		return
	}

	formData := req.Request.Form

	appID := formData.Get("ApplicationID")
	id := formData.Get("CustomerGroupID")

	version := formData.Get("version")
	page := formData.Get("page")
	pageSize := formData.Get("pageSize")

	if "" == appID {
		blog.Error("getContentByCustomerGroupID error: param ApplicationID is empty!input:%#v,rid:%s", formData, srvData.rid)
		converter.RespFailV2(common.CCErrCommParamsNeedSet, defErr.Errorf(common.CCErrCommParamsNeedSet, "ApplicationID").Error(), resp)
		return
	}

	if "" == id {
		blog.Error("getContentByCustomerGroupID error: param CustomerGroupID is empty!input:%#v,rid:%s", formData, srvData.rid)
		converter.RespFailV2(common.CCErrCommParamsNeedSet, defErr.Errorf(common.CCErrCommParamsNeedSet, "CustomerGroupID").Error(), resp)
		return
	}

	name, nameErr := s.GetNameByID(srvData.ctx, appID, id, srvData)
	if nil != nameErr {
		blog.Errorf("getContentByCustomerGroupID error: get CustomerGroup name is error! err:%s,input:%#v,rid:%s", nameErr, formData, srvData.rid)
		converter.RespFailV2Error(err, resp)
		return
	}

	skip := "0"
	if "1" == version {
		intPage, _ := util.GetIntByInterface(page)
		intPageSize, _ := util.GetIntByInterface(pageSize)
		if intPage > 0 {
			intPage--
		} else {
			page = "1"
		}
		if 0 >= intPageSize {
			intPageSize = 20
		}
		skip = strconv.Itoa(intPage * intPageSize)
		pageSize = strconv.Itoa(intPageSize)

	} else {
		pageSize = strconv.Itoa(common.BKNoLimit)
	}

	result, err := s.CoreAPI.HostServer().GetUserCustomQueryResult(srvData.ctx, appID, id, skip, pageSize, srvData.header)
	if nil != err {
		blog.Errorf("GetUserCustomQueryResult http do error.err:%v,input:%#v,rid:%s", err, formData, srvData.rid)
		converter.RespFailV2(common.CCErrCommHTTPDoRequestFailed, defErr.Error(common.CCErrCommHTTPDoRequestFailed).Error(), resp)
		return
	}
	if !result.Result {
		blog.Errorf("GetUserCustomQueryResult http reply error.reply:%#v,input:%#v,rid:%s", result, formData, srvData.rid)
		converter.RespFailV2(result.Code, result.ErrMsg, resp)
		return
	}
	//translate cmdb v3 to v2 api result
	list, total, err := converter.ResToV2ForCustomerGroupResult(result.Result, result.ErrMsg, result.Data)

	//translate cmdb v3 to v2 api result error,
	if err != nil {
		blog.Errorf("getContentByCustomerGroupID  %v", result)
		converter.RespFailV2(common.CCErrCommReplyDataFormatError, defErr.Error(common.CCErrCommReplyDataFormatError).Error(), resp)
		return
	}
	if 0 == len(list) {
		list = make([]common.KvMap, 0)
	}

	if "1" == version {
		ret := make(common.KvMap)
		ret["list"] = list
		ret["total"] = total
		ret["page"] = page
		ret["pageSize"] = pageSize
		ret["GroupName"] = name

		converter.RespSuccessV2(ret, resp)
	} else {
		converter.RespSuccessV2(list, resp)

	}

}

func (s *service) GetNameByID(ctx context.Context, appID, id string, srvData *srvComm) (string, ccError.CCError) {

	result, err := s.CoreAPI.HostServer().GetUserCustomQueryDetail(ctx, appID, id, srvData.header)
	//http request error
	if err != nil {
		blog.Errorf("GetNameByID error:%v", err)
		return "", srvData.ccErr.Error(common.CCErrCommHTTPDoRequestFailed)
	}

	if result.Result {
		name, _ := result.Data["name"].(string)
		return name, nil
	} else {
		return "", srvData.ccErr.New(result.Code, result.ErrMsg)
	}

}

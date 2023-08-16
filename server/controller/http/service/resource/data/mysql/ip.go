/**
 * Copyright (c) 2023 Yunshan Networks
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mysql

import (
	ctrlcommon "github.com/deepflowio/deepflow/server/controller/common"
	"github.com/deepflowio/deepflow/server/controller/http/service/resource/common"
)

type IP struct {
	DataProvider
	dataTool *IPToolData
}

type IPToolData struct {
	wanIP *WANIP
	lanIP *LANIP
}

func NewIP() *IP {
	dp := &IP{
		newDataProvider(ctrlcommon.RESOURCE_TYPE_IP_EN),
		&IPToolData{wanIP: NewWANIP(), lanIP: NewLANIP()},
	}
	dp.setGenerator(dp)
	return dp
}

func (p *IP) generate() ([]common.ResponseElem, error) {
	wanIP, err := p.dataTool.wanIP.generate()
	if err != nil {
		return nil, err
	}
	lanIP, err := p.dataTool.lanIP.generate()
	if err != nil {
		return nil, err
	}
	return mergeResponses(wanIP, lanIP), nil
}

func mergeResponses(resp1 []common.ResponseElem, resp2 []common.ResponseElem) []common.ResponseElem {
	merged := append(resp1, resp2...)
	return merged
}

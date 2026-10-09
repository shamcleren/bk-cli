/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - bk-cli (BlueKing - Cli) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

package monitor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/bk-cli/internal/output"
)

var _ = Describe("response number precision", func() {
	DescribeTable(
		"preserves nested and top-level JSON numbers through envelope output",
		func(status int, raw string) {
			resp := &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(raw)),
			}
			data, err := parseQueryResponse(resp)
			Expect(err).NotTo(HaveOccurred())
			env := output.APIResponse(status, nil, data)
			Expect(env.OK).To(Equal(status >= 200 && status < 300))
			var out bytes.Buffer
			Expect(env.WriteJSON(&out)).To(Succeed())
			var emitted map[string]json.RawMessage
			Expect(json.Unmarshal(out.Bytes(), &emitted)).To(Succeed())
			var compact bytes.Buffer
			Expect(json.Compact(&compact, emitted["data"])).To(Succeed())
			Expect(compact.String()).To(Equal(raw))
		},
		Entry(
			"nested success",
			200,
			`{"id":9007199254740993,"nested":{"values":[18446744073709551615,-9223372036854775809,0.123456789012345678901,1.234567890123456789e+100]}}`,
		),
		Entry("nested error", 502, `{"error":{"id":18446744073709551615,"values":[9007199254740993]}}`),
		Entry("top-level integer", 200, `18446744073709551615`),
		Entry("top-level array", 200, `[9007199254740993,-9223372036854775809,1.25]`),
	)

	DescribeTable("keeps malformed and trailing response bodies as text",
		func(raw string) {
			value, err := parseQueryResponse(&http.Response{Body: io.NopCloser(strings.NewReader(raw))})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(raw))
		},
		Entry("two objects", `{"id":1}{"id":2}`),
		Entry("trailing text", `{"id":9007199254740993} broken`),
		Entry("invalid number", `{"id":01}`),
	)
})

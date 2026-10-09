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

package api

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// normalizeResponseNumbers retains float64 for ordinary response numbers while
// keeping json.Number wherever float conversion could change the emitted value.
func normalizeResponseNumbers(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			typed[key] = normalizeResponseNumbers(item)
		}
	case []any:
		for index, item := range typed {
			typed[index] = normalizeResponseNumbers(item)
		}
	case json.Number:
		return normalizeResponseNumber(typed)
	}
	return value
}

func normalizeResponseNumber(number json.Number) any {
	const maxSafeInteger = 1<<53 - 1
	f, err := number.Float64()
	if err != nil || math.Abs(f) > maxSafeInteger {
		return number
	}
	raw := number.String()
	if !strings.ContainsAny(raw, ".eE") {
		return f
	}
	// Avoid building enormous denominators for underflowed exponents.
	if f == 0 {
		mantissa, _, _ := strings.Cut(strings.ToLower(raw), "e")
		if strings.ContainsAny(mantissa, "123456789") {
			return number
		}
		return f
	}
	original, ok := new(big.Rat).SetString(raw)
	if !ok {
		return number
	}
	roundTrip, ok := new(big.Rat).SetString(strconv.FormatFloat(f, 'g', -1, 64))
	if !ok || original.Cmp(roundTrip) != 0 {
		return number
	}
	return f
}

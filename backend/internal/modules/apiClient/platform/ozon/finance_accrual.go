package ozon

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"slices"
	"strconv"
	"time"

	"autostack/internal/modules/apiClient/platform"
)

const (
	accrualPostingsPath = "/v1/finance/accrual/postings"
	accrualTypesPath    = "/v1/finance/accrual/types"
	accrualPostingLimit = 200
)

type postingAccrualsResponse struct {
	PostingAccruals []struct {
		PostingNumber string           `json:"posting_number"`
		Accruals      []postingAccrual `json:"accruals"`
	} `json:"posting_accruals"`
}

type postingAccrual struct {
	TypeID      int           `json:"type_id"`
	Accrued     accrualMoney  `json:"accrued"`
	Quantity    int32         `json:"quantity"`
	SellerPrice *accrualMoney `json:"seller_price"`
}

type accrualMoney struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}

type accrualTypesResponse struct {
	AccrualTypes []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"accrual_types"`
}

// CommissionData 按发货单汇总的已入账金额。
type CommissionData struct {
	AccrualsForSale         float64
	SaleCommission          float64
	ProcessingAndDelivery   float64
	RefundsAndCancellations float64
	ServicesAmount          float64
	CompensationAmount      float64
	MoneyTransfer           float64
	OthersAmount            float64
	ProfitAmount            float64
	CommissionCurrency      string
}

// GetCommissionsForPostings 查询发货单的已入账金额。没有入账记录的发货单不在结果中。
// https://docs.ozon.ru/api/seller/#operation/GetFinanceAccrualPostings
func (api *FinanceAPI) GetCommissionsForPostings(postingNumbers []string) (map[string]CommissionData, error) {
	result := make(map[string]CommissionData)
	requested := make(map[string]struct{}, len(postingNumbers))
	numbers := make([]string, 0, len(postingNumbers))
	for _, number := range postingNumbers {
		if _, exists := requested[number]; number == "" || exists {
			continue
		}
		requested[number] = struct{}{}
		numbers = append(numbers, number)
	}
	if len(numbers) == 0 {
		return result, nil
	}

	types, err := api.getAccrualTypes()
	if err != nil {
		return nil, err
	}
	unmapped := make(map[int]string)
	for batch := range slices.Chunk(numbers, accrualPostingLimit) {
		request := struct {
			PostingNumbers []string `json:"posting_numbers"`
		}{PostingNumbers: batch}
		body, err := api.requestAccruals(accrualPostingsPath, request)
		if err != nil {
			return nil, fmt.Errorf("获取 Ozon 发货单入账记录失败: %w", err)
		}

		var page postingAccrualsResponse
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("解析 Ozon 发货单入账记录失败: %w", err)
		}
		for _, posting := range page.PostingAccruals {
			if !slices.Contains(batch, posting.PostingNumber) || len(posting.Accruals) == 0 {
				continue
			}
			data := result[posting.PostingNumber]
			for _, accrual := range posting.Accruals {
				name := types[accrual.TypeID]
				recognized, err := data.addAccrual(name, accrual)
				if err != nil {
					return nil, fmt.Errorf("Ozon 发货单 %s 的入账类型 %d 金额无效: %w", posting.PostingNumber, accrual.TypeID, err)
				}
				if !recognized {
					unmapped[accrual.TypeID] = name
				}
			}
			result[posting.PostingNumber] = data
		}
	}
	for id, name := range unmapped {
		log.Printf("[OZON] 入账类型 %d (%s) 已计入其他费用", id, name)
	}
	return result, nil
}

// https://docs.ozon.ru/api/seller/#operation/GetFinanceAccrualTypes
func (api *FinanceAPI) getAccrualTypes() (map[int]string, error) {
	body, err := api.requestAccruals(accrualTypesPath, struct{}{})
	if err != nil {
		return nil, fmt.Errorf("获取 Ozon 入账类型字典失败: %w", err)
	}
	var result accrualTypesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 Ozon 入账类型字典失败: %w", err)
	}
	if len(result.AccrualTypes) == 0 {
		return nil, fmt.Errorf("Ozon 入账类型字典为空")
	}
	types := make(map[int]string, len(result.AccrualTypes))
	for _, accrualType := range result.AccrualTypes {
		types[accrualType.ID] = accrualType.Name
	}
	return types, nil
}

func (data *CommissionData) addAccrual(name string, accrual postingAccrual) (bool, error) {
	// accrued 是此笔入账的总金额，保留平台给出的正负号。
	amount, err := data.readAmount(accrual.Accrued)
	if err != nil {
		return false, err
	}
	data.ProfitAmount += amount

	switch name {
	case "SaleCommission":
		// seller_price 是单价。服务的 accrued 金额已经包含数量，不能再乘一次。
		if accrual.SellerPrice == nil || accrual.Quantity == 0 {
			return false, fmt.Errorf("销售佣金记录缺少单价或有效数量")
		}
		price, err := data.readAmount(*accrual.SellerPrice)
		if err != nil {
			return false, err
		}
		gross := price * float64(accrual.Quantity)
		if math.IsInf(gross, 0) {
			return false, fmt.Errorf("销售金额超出范围")
		}
		data.AccrualsForSale += gross
		data.SaleCommission += amount
		data.ProfitAmount += gross
	case "Logistic", "Shipment", "LastMile", "LastMileCourier", "LastMilePickUpPoint", "Fulfillment",
		"Drop-Off", "Drop-Off Agent", "Pick-Up", "PickUpCourierArrangement", "PickUpCourierDelivery",
		"DeliveryToHandoverPlaceByOzon", "CourierPickUpByOzon", "CourierPickUpReinvoice",
		"RfbsDomesticDelivery", "RfbsGlobalDelivery", "InternationalLogisticDelta", "OzonGlobalLogisticsDelivery",
		"LogisticAgencyFeeToPartner", "LogisticAgencyFeeToOzon", "RfbsGlobalIntermediaryService", "RfbsGlobalPlatformConnectionService",
		"B2C Drop-Off", "B2C Drop-Off Agent", "B2CLogistics", "B2CDeliveryToHandoverPlaceByOzon",
		"B2CCourierClientReinvoice", "B2CPickUpPointClientReinvoice", "ClickAndCollect", "OgrLogistics":
		data.ProcessingAndDelivery += amount
	case "Cancellation", "ClientReturn", "PartialReturn", "BackwardShipment", "ReturnFlowLogistic",
		"PickUpPointReturnAcceptance", "B2CBackwardLogistics", "B2CPickUpPointReturnAcceptance", "RfbsEasyReturn":
		data.RefundsAndCancellations += amount
	case "PayPerClick", "SupplyInbound", "Disposal", "Placements", "ItemPacking", "AcceleratedReviewCollection",
		"PackingFee", "PackageCost", "TemporaryPlacement", "PremiumCashbackIndividualPoints", "CrossDock", "Promotion",
		"BrandCommission", "BrandPromotion", "BrandShelf", "ExternalPromotion", "Marketing", "InternetSiteAdvertising",
		"SocialMediaAdvertising", "DisplayAdvertisingPlacement", "ReturnStorageInTheWarehouse", "TemporaryPlacementsAgent",
		"PointsForReviews", "SaleReview", "ReviewsPin", "PremiumSubscription", "PremiumMembership", "CustomerReviews", "Stencil":
		data.ServicesAmount += amount
	case "Compensation", "ItemCompensation", "B2CInsuranceCompensation":
		data.CompensationAmount += amount
	case "RfbsClientDeliveryCharge", "RfbsBuyerDelivery":
		data.MoneyTransfer += amount
	case "Acquiring":
		data.OthersAmount += amount
	default:
		data.OthersAmount += amount
		return false, nil
	}
	return true, nil
}

// 仅在 Ozon 返回有效 Retry-After 时重试这两个只读财务接口。
func (api *FinanceAPI) requestAccruals(path string, request interface{}) ([]byte, error) {
	const (
		maxRetries = 4
		maxWait    = 10 * time.Second
	)
	for attempt := 0; ; attempt++ {
		response, err := api.client.DoRequest(http.MethodPost, path, request, platform.RequestTypeFinance)
		if err == nil {
			return response.Body, nil
		}
		if attempt >= maxRetries || response == nil || response.StatusCode != http.StatusTooManyRequests {
			return nil, err
		}
		retryAfter := response.Headers.Get("Retry-After")
		seconds, parseErr := strconv.Atoi(retryAfter)
		var delay time.Duration
		if parseErr == nil {
			if seconds < 0 || seconds > int(maxWait/time.Second) {
				return nil, err
			}
			delay = time.Duration(seconds) * time.Second
		} else {
			retryAt, dateErr := http.ParseTime(retryAfter)
			if dateErr != nil {
				return nil, err
			}
			delay = max(0, time.Until(retryAt))
			if delay > maxWait {
				return nil, err
			}
		}
		// 连续限流时逐步延长等待，且至少等待平台指定的时长。
		backoff := min(maxWait, 2*time.Second<<attempt)
		time.Sleep(max(delay, backoff))
	}
}

func (data *CommissionData) readAmount(money accrualMoney) (float64, error) {
	amount, err := money.Amount.Float64()
	if err != nil {
		return 0, err
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("非有限金额")
	}
	if money.Currency == "" {
		return 0, fmt.Errorf("缺少货币")
	}
	if data.CommissionCurrency != "" && data.CommissionCurrency != money.Currency {
		return 0, fmt.Errorf("同一发货单含有多种货币 %s、%s", data.CommissionCurrency, money.Currency)
	}
	data.CommissionCurrency = money.Currency
	return amount, nil
}

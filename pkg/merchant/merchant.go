package merchant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/schanzen/taler-go/pkg/util"
)

type MerchantConfig struct{
	// Default currency
	Currency string `json:"currency"`

	// Supported currencies
	Currencies map[string]util.CurrencySpecification `json:"currencies"`

	// Name
	Name string `json:"name"`

	// Version string
	Version string `json:"version"`
}

type PostOrderRequest struct {
	// The order must at least contain the minimal
	// order detail, but can override all.
	Order MinimalOrderDetail `json:"order"`

	// If set, the backend will then set the refund deadline to the current
	// time plus the specified delay.  If it's not set, refunds will not be
	// possible.
	RefundDelay int64 `json:"refund_delay,omitempty"`

	// Specifies the payment target preferred by the client. Can be used
	// to select among the various (active) wire methods supported by the instance.
	PaymentTarget string `json:"payment_target,omitempty"`

	// Specifies that some products are to be included in the
	// order from the inventory.  For these inventory management
	// is performed (so the products must be in stock) and
	// details are completed from the product data of the backend.
	// FIXME: Not sure we actually need this for now
	//InventoryProducts []MinimalInventoryProduct `json:"inventory_products,omitempty"`

	// Specifies a lock identifier that was used to
	// lock a product in the inventory.  Only useful if
	// inventory_products is set.  Used in case a frontend
	// reserved quantities of the individual products while
	// the shopping cart was being built.  Multiple UUIDs can
	// be used in case different UUIDs were used for different
	// products (i.e. in case the user started with multiple
	// shopping sessions that were combined during checkout).
	LockUuids []string `json:"lock_uuids,omitempty"`

	// Should a token for claiming the order be generated?
	// False can make sense if the ORDER_ID is sufficiently
	// high entropy to prevent adversarial claims (like it is
	// if the backend auto-generates one). Default is 'true'.
	CreateToken bool `json:"create_token,omitempty"`
}

type MinimalOrderDetail struct {
	// Amount to be paid by the customer.
	Amount string `json:"amount"`

	// Short summary of the order.
	Summary string `json:"summary"`

	// See documentation of fulfillment_url in ContractTerms.
	// Either fulfillment_url or fulfillment_message must be specified.
	FulfillmentUrl string `json:"fulfillment_url"`
}

// NOTE: Part of the above but optional
type FulfillmentMetadata struct {
	// See documentation of fulfillment_url in ContractTerms.
	// Either fulfillment_url or fulfillment_message must be specified.
	FulfillmentUrl string `json:"fulfillment_url,omitempty"`

	// See documentation of fulfillment_message in ContractTerms.
	// Either fulfillment_url or fulfillment_message must be specified.
	FulfillmentMessage string `json:"fulfillment_message,omitempty"`
}

type PostOrderResponse struct {
	// Order ID of the response that was just created.
	OrderId string `json:"order_id"`
}

type PostOrderResponseToken struct {
	// Token that authorizes the wallet to claim the order.
	// Provided only if "create_token" was set to 'true'
	// in the request.
	Token string
}

type CheckPaymentStatusResponse struct {
	// Status of the order
	OrderStatus string `json:"order_status"`
}

type CheckPaymentPaytoResponse struct {
	// Status of the order
	TalerPayUri string `json:"taler_pay_uri"`
}

type Merchant struct {

	// The host of this merchant
	BaseUrlPrivate string

	// The access token to use for the private API
	AccessToken string
}

func NewMerchant(merchBaseUrlPrivate string, merchAccessToken string) Merchant {
	return Merchant{
		BaseUrlPrivate: merchBaseUrlPrivate,
		AccessToken:    merchAccessToken,
	}
}

type PaymentStatus string

const (
	OrderStatusUnknown = ""
	OrderPaid = "paid"
	OrderUnpaid = "unpaid"
	OrderClaimed = "claimed"
)

func (m *Merchant) IsOrderPaid(orderId string) (int, PaymentStatus, string, error) {
	var orderPaidResponse CheckPaymentStatusResponse
	var paytoResponse CheckPaymentPaytoResponse
	client := &http.Client{}
	req, _ := http.NewRequest("GET", m.BaseUrlPrivate+"/private/orders/"+orderId, nil)
	req.Header.Set("Authorization", "Bearer secret-token:"+m.AccessToken)
	resp, err := client.Do(req)
	fmt.Println(req)
	if nil != err {
		return resp.StatusCode, OrderStatusUnknown, "", err
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		message := fmt.Sprintf("Expected response code %d. Got %d", http.StatusOK, resp.StatusCode)
		return resp.StatusCode, OrderStatusUnknown, "", errors.New(message)
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, OrderStatusUnknown, "", err
	}
	err = json.NewDecoder(bytes.NewReader(respData)).Decode(&orderPaidResponse)
	if err != nil {
		return resp.StatusCode, OrderStatusUnknown, "", err
	}
	if orderPaidResponse.OrderStatus == "unpaid" {
		err = json.NewDecoder(bytes.NewReader(respData)).Decode(&paytoResponse)
		return resp.StatusCode, PaymentStatus(orderPaidResponse.OrderStatus), paytoResponse.TalerPayUri, err
	}
	return resp.StatusCode, PaymentStatus(orderPaidResponse.OrderStatus), "", nil
}

func (m *Merchant) GetConfig() (*MerchantConfig, error) {
	var configResponse MerchantConfig
	client := &http.Client{}
	req, _ := http.NewRequest("GET", m.BaseUrlPrivate+"/config", nil)
	resp, err := client.Do(req)
	fmt.Println(req)
	if nil != err {
		return nil, err
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		message := fmt.Sprintf("Expected response code %d. Got %d", http.StatusOK, resp.StatusCode)
		return nil, errors.New(message)
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(bytes.NewReader(respData)).Decode(&configResponse)
	if err != nil {
		return nil, err
	}
	return &configResponse, nil
}

func (m *Merchant) AddNewOrder(cost util.Amount, summary string, fulfillment_url string) (string, error) {
	var newOrder PostOrderRequest
	var orderDetail MinimalOrderDetail
	var orderResponse PostOrderResponse
	orderDetail.Amount = cost.String()
	// FIXME get from cfg
	orderDetail.Summary = summary
	orderDetail.FulfillmentUrl = fulfillment_url
	newOrder.Order = orderDetail
	reqString, err := json.Marshal(newOrder)
	if nil != err {
		return "", err
	}
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, m.BaseUrlPrivate+"/private/orders", bytes.NewReader(reqString))
	req.Header.Set("Authorization", "Bearer secret-token:"+m.AccessToken)
	resp, err := client.Do(req)

	if nil != err {
		return "", err
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		message := fmt.Sprintf("Expected response code %d. Got %d. With request %s", http.StatusOK, resp.StatusCode, reqString)
		return "", errors.New(message)
	}
	err = json.NewDecoder(resp.Body).Decode(&orderResponse)
	return orderResponse.OrderId, err
}

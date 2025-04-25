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

type MerchantConfig struct {
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
	Order CommonOrder `json:"order"`

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

type CommonOrder struct {
	// Total price for the transaction. The exchange will subtract deposit
	// fees from that amount before transferring it to the merchant.
	Amount string `json:"amount"`

	// Maximum total deposit fee accepted by the merchant for this contract.
	// Overrides defaults of the merchant instance.
	MaxFee string `json:"max_fee,omitempty"`

	// Human-readable description of the whole purchase.
	Summary string

	// Map from IETF BCP 47 language tags to localized summaries.
	SummaryI18n string `json:"summary_i18n,omitempty"`

	// Unique identifier for the order. Only characters
	// allowed are "A-Za-z0-9" and ".:_-".
	// Must be unique within a merchant instance.
	// For merchants that do not store proposals in their DB
	// before the customer paid for them, the order_id can be used
	// by the frontend to restore a proposal from the information
	// encoded in it (such as a short product identifier and timestamp).
	OrderId string `json:"order_id,omitempty"`

	// URL where the same contract could be ordered again (if
	// available). Returned also at the public order endpoint
	// for people other than the actual buyer (hence public,
	// in case order IDs are guessable).
	PublicReorderUrl string `json:"public_reorder_url,omitempty"`

	// See documentation of fulfillment_url field in ContractTerms.
	// Either fulfillment_url or fulfillment_message must be specified.
	// When creating an order, the fulfillment URL can
	// contain ${ORDER_ID} which will be substituted with the
	// order ID of the newly created order.
	FulfillmentUrl string `json:"fulfillment_url,omitempty"`

	// See documentation of fulfillment_message in ContractTerms.
	// Either fulfillment_url or fulfillment_message must be specified.
	FulfillmentMessage string `json:"fulfillment_message,omitempty"`

	// Map from IETF BCP 47 language tags to localized fulfillment
	// messages.
	FulfillmentMessageI18n string `json:"fulfillment_message_i18n,omitempty"`

	// Minimum age the buyer must have to buy.
	MinimumAge *uint64 `json:"minimum_age,omitempty"`

	// List of products that are part of the purchase.
	//products?: Product[];

	// Time when this contract was generated. If null, defaults to current
	// time of merchant backend.
	Timestamp *uint64 `json:"timestamp,omitempty"`

	// After this deadline has passed, no refunds will be accepted.
	// Overrides deadline calculated from refund_delay in
	// PostOrderRequest.
	RefundDeadline *uint64 `json:"refund_deadline,omitempty"`

	// After this deadline, the merchant won't accept payments for the contract.
	// Overrides deadline calculated from default pay delay configured in
	// merchant backend.
	PayDeadline *uint64 `json:"pay_deadline,omitempty"`

	// Transfer deadline for the exchange. Must be in the deposit permissions
	// of coins used to pay for this order.
	// Overrides deadline calculated from default wire transfer delay
	// configured in merchant backend. Must be after refund deadline.
	WireTransferDeadline *uint64 `json:"wire_transfer_deadline,omitempty"`

	// Base URL of the (public!) merchant backend API.
	// Must be an absolute URL that ends with a slash.
	// Defaults to the base URL this request was made to.
	MerchantBaseUrl string `json:"merchant_base_url,omitempty"`

	// Delivery location for (all!) products.
	//DeliveryLocation?: Location;

	// Time indicating when the order should be delivered.
	// May be overwritten by individual products.
	// Must be in the future.
	DeliveryDate *uint64 `json:"delivery_deadline,omitempty"`

	// See documentation of auto_refund in ContractTerms.
	// Specifies for how long the wallet should try to get an
	// automatic refund for the purchase.
	AutoRefund *uint64 `json:"auto_refund,omitempty"`

	// Extra data that is only interpreted by the merchant frontend.
	// Useful when the merchant needs to store extra information on a
	// contract without storing it separately in their database.
	// Must really be an Object (not a string, integer, float or array).
	Extra string
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
	OrderPaid          = "paid"
	OrderUnpaid        = "unpaid"
	OrderClaimed       = "claimed"
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

func (m *Merchant) CreateOrder(order CommonOrder) (string, error) {
	var newOrder PostOrderRequest
	var orderResponse PostOrderResponse
	newOrder.Order = order
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

func (m *Merchant) AddNewOrder(cost util.Amount, summary string, fulfillment_url string) (string, error) {
	var newOrder PostOrderRequest
	var orderDetail CommonOrder
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

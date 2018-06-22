package http

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/gin-gonic/gin"
)

//SetRebalanceQuadratic set pending rebalance quadratic equation
//input data follow json: {"data":{"KNC": {"a": 0.7, "b": 1.2, "c": 1.3}}}
func (h *HTTPServer) SetRebalanceQuadratic(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"value"}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	var rq metric.RebalanceQuadraticRequest
	if err := json.Unmarshal(value, &rq); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := rq.IsValid(); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := h.metric.StorePendingRebalanceQuadratic(value); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

//GetPendingRebalanceQuadratic return currently pending config for rebalance quadratic equation
//if there is no pending equation return success false
func (h *HTTPServer) GetPendingRebalanceQuadratic(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := h.metric.GetPendingRebalanceQuadratic()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

//ConfirmRebalanceQuadratic confirm configuration for current pending config for rebalance quadratic equation
func (h *HTTPServer) ConfirmRebalanceQuadratic(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := h.metric.ConfirmRebalanceQuadratic(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

//RejectRebalanceQuadratic reject pending configuration for rebalance quadratic function
func (h *HTTPServer) RejectRebalanceQuadratic(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := h.metric.RemovePendingRebalanceQuadratic()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

//GetRebalanceQuadratic return current confirmed rebalance quadratic equation
func (h *HTTPServer) GetRebalanceQuadratic(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := h.metric.GetRebalanceQuadratic()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

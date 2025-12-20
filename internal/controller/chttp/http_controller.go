package chttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/MaKcm14/price-service/internal/entities/dto"
	"github.com/MaKcm14/price-service/internal/services"
	"github.com/MaKcm14/price-service/internal/services/filter"
	"github.com/MaKcm14/price-service/pkg/entities"

	_ "github.com/MaKcm14/price-service/docs"
)

// Controller handles the clients' requests.
type Controller struct {
	contr  *echo.Echo
	logger *slog.Logger
	filter filter.Filter
	valid  validator
	conf   ControllerConfig
}

func NewController(
	conf ControllerConfig,
	contr *echo.Echo,
	logger *slog.Logger,
	filter filter.Filter,
) Controller {
	return Controller{
		contr:  contr,
		logger: logger,
		filter: filter,
		conf:   conf,
	}
}

// Run configures and starts the http-server.
func (c *Controller) Run(socket string) {
	defer c.logger.Info("the http-server was stopped")
	defer c.contr.Close()

	c.logger.Info("configuring and starting the http-server begun")

	c.configController()
	if err := c.contr.Start(socket); err != nil {
		serverErr := fmt.Errorf("http-server wasn't started or was closed: %v", err)
		c.logger.Error(serverErr.Error())
		panic(serverErr)
	}
}

// GracefulStop defines the logic of service's gracefully shutdown.
func (c *Controller) GracefulStop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := c.contr.Shutdown(ctx); err != nil {
		c.logger.Error(fmt.Sprintf("error of gracefully shutdown the server: %v", err))
	}

	c.logger.Info("server was gracefully stopped")
}

// configController configurates the controller by setting the middlewares and paths.
func (c *Controller) configController() {
	c.configMW()
	c.configPaths()
}

// configPaths configurates the controller's paths and methods.
func (c *Controller) configPaths() {
	c.contr.GET("/products/filter/price/price-range", c.handlePriceRangeRequest)
	c.contr.GET("/products/filter/price/best-price", c.handleBestPriceRequest)
	c.contr.GET("/products/filter/price/exact-price", c.handleExactPriceRequest)
	c.contr.GET("/products/filter/markets", c.handleMarketsRequest)

	c.contr.POST("/products/filter/price/best-price/async", c.handleBestPriceAsyncRequest)

	c.contr.GET("/swagger/*", echoSwagger.WrapHandler)
	c.contr.GET("/api/markets", c.handleMarkets)
}

// configMW configurates the controller's middleware.
func (c *Controller) configMW() {
	c.contr.Use(middleware.BodyLimit("600K"))
	c.contr.Use(middleware.Gzip())

	c.contr.HTTPErrorHandler = func(err error, ctx echo.Context) {
		if errHttp, flagCheck := err.(*echo.HTTPError); flagCheck {
			if errHttp.Code == http.StatusNotFound {
				c.logger.Warn("the wrong request path was got")
				ctx.JSON(http.StatusNotFound, ResponseErr{ErrRequestPath.Error()})
			} else {
				c.logger.Warn(fmt.Sprintf("%v", err))
				ctx.JSON(http.StatusBadRequest, ResponseErr{ErrRequest.Error()})
			}
			return
		}

		if err != nil {
			c.logger.Warn(fmt.Sprintf("%v", err))
			ctx.JSON(http.StatusInternalServerError, ResponseErr{ErrServerHandling.Error()})
		}
	}
}

// setBasicHeaders sets the headers that are common for every successfull response.
func (c *Controller) setBasicHeaders(ctx echo.Context, buf []byte) {
	ctx.Response().Header().Add("Connection", "keep-alive")
	ctx.Response().Header().Add("Content-Language", "en, ru")
	ctx.Response().Header().Add("Content-Length", fmt.Sprintf("%d", len(buf)+1))
}

// handlePriceRangeRequest defines the logic of the handling the filter-by-price-down-up requests.
//
//	@summary		price range filtering
//	@description	this endpoint provides filtering products from marketplaces with specified price range
//	@tags			Price-Filters
//	@produce		json
//
//	@param			query		query		[]string	true	"the exact query string"								collectionFormat(ssv)	minLength(1)	example(iphone+11)
//	@param			price_down	query		integer		true	"the price range's lower bound: less than price_up"		minimum(0)
//	@param			price_up	query		integer		true	"the price range's upper bound: more than price_down"	minimum(1)
//	@param			markets		query		[]string	true	"the list of the markets using for search"				Enums(wildberries, megamarket)				collectionFormat(ssv)	minLength(1)	example(megamarket+wildberries)
//	@param			sample		query		integer		false	"the num of products' sample"							minimum(1)									default(1)
//	@param			sort		query		string		false	"the type of products' sample sorting"					Enums(popular, pricedown, priceup, newly)	default(popular)
//	@param			no-image	query		integer		false	"the flag that defines 'Should image links be parsed?'"	Enums(0, 1)									default(1)
//	@param			amount		query		string		false	"the amount of the products in response's sample"		Enums(min, max)								default(min)
//
//
//	@success		200			{object}	chttp.ProductResponse
//	@failure		400			{object}	chttp.ResponseErr
//	@failure		500			{object}	chttp.ResponseErr
//	@failure		502			{object}	chttp.ResponseErr
//	@router			/products/filter/price/price-range [get]
func (c *Controller) handlePriceRangeRequest(ctx echo.Context) error {
	const filterType = "price-range-filter"

	requestInfo, err := c.valid.validProductRequest(ctx,
		c.valid.validQuery,
		c.valid.validMarkets,
		c.valid.validAmount,
		c.valid.validSample,
		c.valid.validSort,
		c.valid.validNoImage,
	)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{ErrRequestInfo.Error()})
	}
	priceDown, _ := strconv.Atoi(ctx.QueryParam("price_down"))
	priceUp, _ := strconv.Atoi(ctx.QueryParam("price_up"))

	if priceDown < 0 || priceUp <= 0 || priceUp < priceDown {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, ErrRequestInfo))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{ErrRequestInfo.Error()})
	}

	requestInfo.PriceRange = dto.PriceRangeRequest{
		PriceDown: priceDown,
		PriceUp:   priceUp,
	}

	products, err := c.filter.FilterByPriceRange(ctx, requestInfo)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))

		if errors.Is(err, services.ErrGettingProducts) {
			return ctx.JSON(http.StatusBadGateway, ResponseErr{ErrExternalServer.Error()})
		}

		return ctx.JSON(http.StatusInternalServerError, ResponseErr{ErrServerHandling.Error()})
	}

	response := NewProductResponse(products)
	buf, _ := json.Marshal(response)

	c.setBasicHeaders(ctx, buf)

	ctx.Response().Header().Add("Content-Type", "application/json; charset=utf-8")
	ctx.Response().Header().Add("Cache-Control", "public, max-age=43200")

	return ctx.JSON(http.StatusOK, response)
}

// handleBestPriceRequest defines the logic of the handling the filter-by-minimal-price requests.
//
//	@summary		best price filtering
//	@description	this endpoint provides filtering products from marketplaces with the best and minimum price
//	@tags			Price-Filters
//	@produce		json
//
//	@param			query		query		[]string	true	"the exact query string"								collectionFormat(ssv)						minLength(1)			example(iphone+11)
//	@param			markets		query		[]string	true	"the list of the markets using for search"				Enums(wildberries, megamarket)				collectionFormat(ssv)	minLength(1)	example(megamarket+wildberries)
//	@param			sample		query		integer		false	"the num of products' sample"							minimum(1)									default(1)
//	@param			sort		query		string		false	"the type of products' sample sorting"					Enums(popular, pricedown, priceup, newly)	default(popular)
//	@param			no-image	query		integer		false	"the flag that defines 'Should image links be parsed?'"	Enums(0, 1)									default(1)
//	@param			amount		query		string		false	"the amount of the products in response's sample"		Enums(min, max)								default(min)
//
//
//	@success		200			{object}	chttp.ProductResponse
//	@failure		400			{object}	chttp.ResponseErr
//	@failure		500			{object}	chttp.ResponseErr
//	@failure		502			{object}	chttp.ResponseErr
//	@router			/products/filter/price/best-price [get]
func (c *Controller) handleBestPriceRequest(ctx echo.Context) error {
	const filterType = "best-price-filter"

	requestInfo, err := c.valid.validProductRequest(ctx,
		c.valid.validQuery,
		c.valid.validMarkets,
		c.valid.validAmount,
		c.valid.validSample,
		c.valid.validNoImage,
	)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{ErrRequestInfo.Error()})
	}

	products, err := c.filter.FilterByBestPrice(ctx, requestInfo)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))

		if errors.Is(err, services.ErrGettingProducts) {
			return ctx.JSON(http.StatusBadGateway, ResponseErr{ErrExternalServer.Error()})
		}

		return ctx.JSON(http.StatusInternalServerError, ResponseErr{ErrServerHandling.Error()})
	}

	response := NewProductResponse(products)
	buf, _ := json.Marshal(response)

	c.setBasicHeaders(ctx, buf)

	ctx.Response().Header().Add("Content-Type", "application/json; charset=utf-8")
	ctx.Response().Header().Add("Cache-Control", "public, max-age=43200")

	return ctx.JSON(http.StatusOK, response)
}

// handleExactPriceRequest defines the logic of the handling the filter-by-set-price requests.
//
//	@summary		exact price filtering
//	@description	this endpoint provides filtering products from marketplaces with price in range (exact-price, exact-price * 1.05 (+5%))
//	@tags			Price-Filters
//	@produce		json
//
//	@param			query		query		[]string	true	"the exact query string"									collectionFormat(ssv)	minLength(1)	example(iphone+11)
//	@param			price		query		integer		true	"the value of exact price"									minimum(1)
//	@param			markets		query		[]string	true	"the list of the markets using for search"					Enums(wildberries, megamarket)				collectionFormat(ssv)	minLength(1)	example(megamarket+wildberries)
//	@param			sample		query		integer		false	"the num of products' sample"								minimum(1)									default(1)
//	@param			sort		query		string		false	"the type of products' sample sorting"						Enums(popular, pricedown, priceup, newly)	default(popular)
//	@param			no-image	query		integer		false	"the flag that defines 'Should image links be parsed??'"	Enums(0, 1)									default(1)
//	@param			amount		query		string		false	"the amount of the products in response's sample"			Enums(min, max)								default(min)
//
//
//	@success		200			{object}	chttp.ProductResponse
//	@failure		400			{object}	chttp.ResponseErr
//	@failure		500			{object}	chttp.ResponseErr
//	@failure		502			{object}	chttp.ResponseErr
//	@router			/products/filter/price/exact-price [get]
func (c *Controller) handleExactPriceRequest(ctx echo.Context) error {
	const filterType = "exact-price-filter"

	requestInfo, err := c.valid.validProductRequest(ctx,
		c.valid.validQuery,
		c.valid.validMarkets,
		c.valid.validAmount,
		c.valid.validSample,
		c.valid.validSort,
		c.valid.validNoImage,
	)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{ErrRequestInfo.Error()})
	}

	exactPrice, _ := strconv.Atoi(ctx.QueryParam("price"))

	if exactPrice <= 0 {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, ErrRequestInfo))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{ErrRequestInfo.Error()})
	}

	requestInfo.ExactPrice = exactPrice

	products, err := c.filter.FilterByExactPrice(ctx, requestInfo)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))

		if errors.Is(err, services.ErrGettingProducts) {
			return ctx.JSON(http.StatusBadGateway, ResponseErr{ErrExternalServer.Error()})
		}

		return ctx.JSON(http.StatusInternalServerError, ResponseErr{ErrServerHandling.Error()})
	}
	response := NewProductResponse(products)
	buf, _ := json.Marshal(response)

	c.setBasicHeaders(ctx, buf)

	ctx.Response().Header().Add("Content-Type", "application/json; charset=utf-8")
	ctx.Response().Header().Add("Cache-Control", "public, max-age=43200")

	return ctx.JSON(http.StatusOK, response)
}

// handleMarketsRequest defines the logic of the handling the filter-by-markets requests.
//
//	@summary		common filtering
//	@description	this endpoint provides filtering products from marketplaces without any specified filtration
//	@tags			Common-Filters
//	@produce		json
//
//	@param			query		query		[]string	true	"the exact query string"								collectionFormat(ssv)						minLength(1)			example(iphone+11)
//	@param			markets		query		[]string	true	"the list of the markets using for search"				Enums(wildberries, megamarket)				collectionFormat(ssv)	minLength(1)	example(megamarket+wildberries)
//	@param			sample		query		integer		false	"the num of products' sample"							minimum(1)									default(1)
//	@param			sort		query		string		false	"the type of products' sample sorting"					Enums(popular, pricedown, priceup, newly)	default(popular)
//	@param			no-image	query		integer		false	"the flag that defines 'Should image links be parsed?'"	Enums(0, 1)									default(1)
//	@param			amount		query		string		false	"the amount of the products in response's sample"		Enums(min, max)								default(min)
//
//
//	@success		200			{object}	chttp.ProductResponse
//	@failure		400			{object}	chttp.ResponseErr
//	@failure		500			{object}	chttp.ResponseErr
//	@failure		502			{object}	chttp.ResponseErr
//	@router			/products/filter/markets [get]
func (c *Controller) handleMarketsRequest(ctx echo.Context) error {
	const filterType = "markets-filter"

	requestInfo, err := c.valid.validProductRequest(ctx,
		c.valid.validQuery,
		c.valid.validMarkets,
		c.valid.validAmount,
		c.valid.validSample,
		c.valid.validSort,
		c.valid.validNoImage,
	)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{err.Error()})
	}

	products, err := c.filter.FilterByMarkets(ctx, requestInfo)

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, ErrServerHandling))

		if errors.Is(err, services.ErrGettingProducts) {
			return ctx.JSON(http.StatusBadGateway, ResponseErr{ErrExternalServer.Error()})
		}

		return ctx.JSON(http.StatusInternalServerError, ResponseErr{ErrServerHandling.Error()})
	}

	response := NewProductResponse(products)
	buf, _ := json.Marshal(response)

	c.setBasicHeaders(ctx, buf)

	ctx.Response().Header().Add("Content-Type", "application/json; charset=utf-8")
	ctx.Response().Header().Add("Cache-Control", "public, max-age=43200")

	return ctx.JSON(http.StatusOK, response)
}

// handleMarkets defines the logic of handling the markets request:
// it returns the markets, that can be use for the search.
//
//	@summary		markets getting
//	@description	this endpoint provides getting the current markets that are supported by the service
//	@tags			Service-Info
//	@produce		json
//
//	@success		200	{object}	entities.SupportedMarkets
//	@router			/api/markets [get]
func (c *Controller) handleMarkets(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, entities.GetSupportedMarkets())
}

// handleBestPriceAsyncRequest defines the logic of handling the best-price request
// with the async processing.
//
//	@summary		async best price filtering
//	@description	this endpoint provides filtering products from marketplaces with the best and minimum price in async mode
//	@tags			Price-Filters
//	@produce		json
//
//	@param			query		query	[]string			true	"the exact query string"								collectionFormat(ssv)						minLength(1)			example(iphone+11)
//	@param			markets		query	[]string			true	"the list of the markets using for search"				Enums(wildberries, megamarket)				collectionFormat(ssv)	minLength(1)	example(megamarket+wildberries)
//	@param			sample		query	integer				false	"the num of products' sample"							minimum(1)									default(1)
//	@param			sort		query	string				false	"the type of products' sample sorting"					Enums(popular, pricedown, priceup, newly)	default(popular)
//	@param			no-image	query	integer				false	"the flag that defines 'Should image links be parsed?'"	Enums(0, 1)									default(1)
//	@param			amount		query	string				false	"the amount of the products in response's sample"		Enums(min, max)								default(min)
//	@param			request		body	chttp.extraHeaders	true	"the headers that need to be included into the async response"
//
//	@success		200
//	@failure		400	{object}	chttp.ResponseErr
//	@router			/products/filter/price/best-price/async [post]
func (c *Controller) handleBestPriceAsyncRequest(ctx echo.Context) error {
	const filterType = "async-best-price-filter"

	if !c.conf.AsyncMode {
		return ctx.JSON(http.StatusServiceUnavailable, ResponseErr{
			ErrRequestMode.Error(),
		})
	}

	requestInfo, err := c.valid.validProductRequest(ctx,
		c.valid.validQuery,
		c.valid.validMarkets,
		c.valid.validAmount,
		c.valid.validSample,
		c.valid.validNoImage,
		c.valid.validExtraHeaders,
	)
	requestInfo.Async = true

	if err != nil {
		c.logger.Warn(fmt.Sprintf("error of the %v: %v", filterType, err))
		return ctx.JSON(http.StatusBadRequest, ResponseErr{err.Error()})
	}

	go c.filter.FilterByBestPriceAsync(ctx, requestInfo)

	return ctx.JSON(http.StatusOK, nil)
}

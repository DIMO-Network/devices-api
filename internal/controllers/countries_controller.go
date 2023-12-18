package controllers

import (
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/gofiber/fiber/v2"
)

type CountriesController struct {
}

func NewCountriesController() CountriesController {
	return CountriesController{}
}

// GetSupportedCountries godoc
// @Description Returns all the supported countries
// @Tags        countries
// @Produce     json
// @Success     200            {object} constants.CountryInfo
// @Router      /countries [get]
func (cc CountriesController) GetSupportedCountries(c *fiber.Ctx) error {
	// read countries json and return it
	return c.JSON(constants.GetCountriesRaw())
}

// GetCountry godoc
// @Description Returns a specific country by 3-letter ISO country code
// @Tags        countries
// @Produce     json
// @Param       countryCode path string true "3-letter country code"
// @Success     200            {object} constants.CountryInfo
// @Failure     404 "country not found with that country code"
// @Failure     400 "invalid country code"
// @Router      /countries/{countryCode} [get]
func (cc CountriesController) GetCountry(c *fiber.Ctx) error {
	countryCode := c.Params("countryCode")
	if len(countryCode) != 3 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid country code: "+countryCode)
	}

	country := constants.FindCountry(countryCode)
	if country == nil {
		return fiber.NewError(fiber.StatusNotFound, "country not found or not supported: "+countryCode)
	}

	return c.JSON(country)
}

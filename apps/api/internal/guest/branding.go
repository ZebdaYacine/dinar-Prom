package guest

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type branding struct {
	Site     string `json:"site"`
	Locale   string `json:"locale"`
	Proverb  string `json:"proverb"`
	LogoURL  string `json:"logo_url"`
	LogoAlt  string `json:"logo_alt"`
	IsActive bool   `json:"is_active"`
}

// RegisterBranding is deliberately small and public-safe. EPIC-08 will replace
// the source with site_branding in PostgreSQL without changing this contract.
func RegisterBranding(router fiber.Router) {
	router.Get("/site/branding", func(ctx *fiber.Ctx) error {
		site := strings.TrimSpace(ctx.Query("site", "dinar_dz"))
		locale := strings.TrimSpace(ctx.Query("locale", "ar-DZ"))
		if site != "dinar_dz" || (locale != "ar-DZ" && locale != "fr-FR") {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "INVALID_BRANDING_QUERY", "message": "site or locale is not supported"}})
		}

		proverb := "كل تخفيضات مدينتك في مكان واحد"
		if locale == "fr-FR" {
			proverb = "Toutes les réductions de votre ville au même endroit"
		}
		return ctx.JSON(fiber.Map{"data": branding{Site: site, Locale: locale, Proverb: proverb, LogoURL: "/brand/dinar-dz-mark.svg", LogoAlt: "Dinar DZ logo", IsActive: true}})
	})
}

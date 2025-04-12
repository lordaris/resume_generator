package security

import (
	"net/http"
)

// HeadersConfig contains configuration options for security headers
type HeadersConfig struct {
	// HSTS determines if Strict-Transport-Security header should be set
	HSTS bool
	// HSTSMaxAge is the max age for HSTS in seconds
	HSTSMaxAge int
	// HSTSIncludeSubdomains determines if includeSubDomains directive is added to HSTS
	HSTSIncludeSubdomains bool
	// HSTSPreload determines if preload directive is added to HSTS
	HSTSPreload bool
	// ContentSecurityPolicy is the Content-Security-Policy header value
	ContentSecurityPolicy string
	// XFrameOptions is the X-Frame-Options header value
	XFrameOptions string
	// XContentTypeOptions is the X-Content-Type-Options header value
	XContentTypeOptions string
}

// DefaultHeadersConfig returns the default configuration for security headers
func DefaultHeadersConfig() HeadersConfig {
	return HeadersConfig{
		HSTS:                  true,
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'self'; object-src 'none'",
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
	}
}

// SecurityHeaders provides middleware for setting security headers
func SecurityHeaders(config HeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set Strict-Transport-Security header
			if config.HSTS {
				var value string
				if config.HSTSMaxAge > 0 {
					value = "max-age=" + http.StatusText(config.HSTSMaxAge)
					if config.HSTSIncludeSubdomains {
						value += "; includeSubDomains"
					}
					if config.HSTSPreload {
						value += "; preload"
					}
				}
				if value != "" {
					w.Header().Set("Strict-Transport-Security", value)
				}
			}

			// Set Content-Security-Policy header
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Set X-Frame-Options header
			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			}

			// Set X-Content-Type-Options header
			if config.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.XContentTypeOptions)
			}

			// Set additional security headers
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
			w.Header().Set("Cache-Control", "no-store, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			next.ServeHTTP(w, r)
		})
	}
}

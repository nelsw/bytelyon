package model

// SameSiteAttribute is exactly what you think it is; Strict, Lax, or None.
type SameSiteAttribute string

type BrowserContextState struct {
	// Cookies to set for context
	Cookies []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		// Either url or domain / path are required. Optional.
		URL *string `json:"url"`
		// For the cookie to apply to all subdomains as well, prefix domain with a dot, like this: ".example.com". Either url
		// or domain / path are required. Optional.
		Domain *string `json:"domain"`
		// Either url or domain / path are required Optional.
		Path *string `json:"path"`
		// Unix time in seconds. Optional.
		Expires *float64 `json:"expires"`
		// Optional.
		HttpOnly *bool `json:"httpOnly"`
		// Optional.
		Secure *bool `json:"secure"`
		// Optional.
		SameSite *SameSiteAttribute `json:"sameSite"`
	} `json:"cookies"`
	// localStorage to set for context
	Origins []struct {
		Origin       string `json:"origin"`
		LocalStorage []struct {
			// Name of the header.
			Name string `json:"name"`
			// Value of the header.
			Value string `json:"value"`
		} `json:"localStorage"`
	} `json:"origins"`
}

package shopify

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/rs/zerolog/log"
)

const shopifyAdmin = "https://msnbic-0w.myshopify.com/admin"
const shopifyAuth = shopifyAdmin + "/oauth/access_token"
const shopifyAPI = shopifyAdmin + "/api/2026-01/graphql.json"

type AccessTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type CreateArticleResponse struct {
	Data struct {
		ArticleCreate struct {
			Article    any `json:"article"`
			UserErrors []struct {
				Code    string   `json:"code"`
				Field   []string `json:"field"`
				Message string   `json:"message"`
			} `json:"userErrors"`
		} `json:"articleCreate"`
	} `json:"data"`
	Extensions struct {
		Cost struct {
			RequestedQueryCost int `json:"requestedQueryCost"`
			ActualQueryCost    int `json:"actualQueryCost"`
			ThrottleStatus     struct {
				MaximumAvailable   float64 `json:"maximumAvailable"`
				CurrentlyAvailable int     `json:"currentlyAvailable"`
				RestoreRate        float64 `json:"restoreRate"`
			} `json:"throttleStatus"`
		} `json:"cost"`
	} `json:"extensions"`
}

func (r *CreateArticleResponse) error() (err error) {
	for _, e := range r.Data.ArticleCreate.UserErrors {
		log.Warn().
			Str("code", e.Code).
			Strs("field", e.Field).
			Msg(e.Message)
		err = errors.Join(err, errors.New(e.Message))
	}
	return
}

func AccessToken() (string, error) {
	return accessToken()
}

func accessToken() (string, error) {

	out, err := https.PostForm(shopifyAuth, map[string][]string{
		"grant_type":    {"client_credentials"},
		"client_id":     {os.Getenv("SHOPIFY_CLIENT_ID")},
		"client_secret": {os.Getenv("SHOPIFY_CLIENT_SECRET")},
	})

	if err != nil {
		return "", err
	}

	var atr AccessTokenResponse
	if err = json.Unmarshal(out, &atr); err != nil {
		return "", err
	}

	return atr.AccessToken, nil
}

// CreateArticle creates a Shopify Article.
// https://shopify.dev/docs/api/admin-graphql/latest/mutations/articleCreate
func CreateArticle(in []byte) (err error) {

	var tkn string
	if tkn, err = accessToken(); err != nil {
		return
	}

	var out []byte
	out, err = https.PostJSON(shopifyAPI, in, map[string]string{
		"Content-Type":           "application/json",
		"X-Shopify-Access-Token": tkn,
	})

	if err != nil {
		return
	}

	var r CreateArticleResponse
	if err = json.Unmarshal(out, &r); err != nil {
		return
	} else if err = r.error(); err != nil {
		return
	}

	return nil
}

/*
Palm Beach International Airport Leads the Way in EV Fire Protection with Revolutionary "Cold Cut Cobra" System
Electric vehicles are transforming the way we drive — but they're also transforming the challenges faced by firefighters across the country. As EV adoption continues to surge, fire rescue teams are grappling with a dangerous reality: electric vehicle fires are fundamentally different from traditional car fires, and they demand entirely new approaches to containment and suppression.

The Growing Challenge of EV Fires
According to Matt Ritter, district chief of Palm Beach County Fire Rescue, EV fires produce a staggering amount of toxic gases that pose serious risks not only to the public but to the firefighters themselves. These gases are so corrosive and hazardous that they can render firefighting gear virtually unusable after a single exposure.

"With EV fires, they produce a large amount of toxic gases, which we found that when it gets on our gear, the gear is almost to the point of garbage. It's almost impossible to clean," Ritter explained.

Even more alarming are the health implications. "Anything that's going to do that, if it's inhaled or gets on the skin of the firefighters, can make them extremely sick to the point that they may not even be able to come back to work," he added.

This underscores the urgent need for better ev fire protection strategies — tools and technologies that can suppress lithium-ion battery fires faster, safer, and more effectively than traditional methods.

A First-of-Its-Kind Solution Takes Flight
Palm Beach International Airport has now become the first airport in the United States to implement the "Cold Cut Cobra" system — a cutting-edge firefighting tool specifically designed to tackle the unique dangers of electric vehicle fires.

Originally developed for the explosive industry, the Cold Cut Cobra has been adapted for firefighting applications with remarkable results. Carl Batchelor, a senior Cobra instructor, described the system's impressive capabilities: "It's an ultra-high pressure system. A system that works at 4,300 PSI, and the water mist comes out at 447 miles an hour."

The system works by piercing through the exterior of an electric vehicle and creating a hole directly in the lithium battery, then flooding it with water. This targeted approach addresses one of the most persistent challenges in EV fire suppression — reignition. Lithium-ion batteries are notorious for reigniting hours or even days after an initial fire is extinguished, making conventional firefighting methods frustratingly inadequate.

Why This Matters for Airports and Parking Structures
One of the most compelling advantages of the Cold Cut Cobra system is its portability and versatility, particularly in environments where traditional fire engines simply can't reach.

"We can use this device in an airport and in any parking garage. It's very difficult, actually impossible, to get a typical fire engine up to the higher floor. We can deploy this at one international airport to go up to the parking garages and take out those car fires, with a minimal disruption to the airport," Ritter explained.

Think about that for a moment. As more travelers arrive at airports in electric vehicles and park in multi-level garages, the risk of an EV fire in a confined, hard-to-reach space grows significantly. Having a deployable, high-pressure system that uses significantly less water than traditional methods isn't just convenient — it's essential.

A Game Changer for EV Fire Protection
District Chief Ritter didn't mince words about the impact of this new technology: "This is a game changer for us. It's a game changer for the citizens of Palm Beach County. It is a game changer financially I think also."

The reduced water usage, the ability to prevent reignition, the portability for multi-story structures, and the enhanced safety for firefighters all add up to a revolutionary step forward in ev fire protection.

The Bigger Picture
As electric vehicle adoption accelerates nationwide, every airport, parking structure, residential community, and commercial property needs to be thinking proactively about lithium-ion battery fire risks. The Cold Cut Cobra is one powerful tool in the arsenal, but comprehensive ev fire protection requires a multi-layered approach — from early detection and suppression systems to protective barriers and fire blankets designed specifically for lithium-ion battery emergencies.

At FireFibers, we're committed to staying at the forefront of this critical conversation. Our lithium-ion fire blankets provide an essential layer of protection for containing and suppressing battery fires in vehicles, storage facilities, and beyond. Because when it comes to EV fire safety, preparation isn't optional — it's imperative.

For the full story on Palm Beach International Airport's groundbreaking adoption of the Cold Cut Cobra system, check out the original article from WPBF.

Interested in learning how FireFibers lithium-ion fire blankets can enhance your ev fire protection strategy? Contact us today.

https://www.wpbf.com/article/florida-palm-beach-international-airport-introduces-new-tool-to-fight-electric-vehicle-fires/70115270
*/

//https://kubrick.htvapps.com/vidthumb/f43a2ce6-b4f1-4585-b843-2ccf2daa707b/984c9bfe-c14a-4da9-a2cc-1672d20b9978.jpg

package dto

import (
	"fmt"
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/util"
)

func TestNewNode(t *testing.T) {
	logs.Init("trace")
	urls := []string{
		"https://firefibers.com",
		"https://firefibers.com/blogs/news",
		"https://firefibers.com/blogs/news/car-burned-in-pasco-fire-involving-lithium-ion-batteries-01kqx4pbapb0ryxhen59w4911t",
		"https://firefibers.com/blogs/news/electric-scooter-showroom-fire-in-lucknow-a-wake-up-call-for-ev-battery-safety-01kqc13mq0vfh5aqnc3bjmhpjp",
		"https://firefibers.com/blogs/news/electric-vehicle-catches-fire-at-irvine-intersection-national-today-01kr5r2b2n23edehtdrjgvybzp",
		"https://firefibers.com/blogs/news/ev-charger-fire-incident-serves-as-important-safety-reminder-01kpwbqy90b8cd1m1ez7mp56s2",
		"https://firefibers.com/blogs/news/ev-fires-dont-play-by-the-old-rules-and-that-changes-everything-01km8eaq000cje3df498cdvy50",
		"https://firefibers.com/blogs/news/ev-fires-rare-but-relentless-what-the-bedford-toll-plaza-incident-teaches-us-about-lithium-ion-battery-safety-01kq0281y0wktg20ekhrea7sqk",
		"https://firefibers.com/blogs/news/fire-crews-tackle-suspected-battery-blaze-at-saanich-landfill",
		"https://firefibers.com/blogs/news/gurgaon-apartments-mandate-removal-of-private-ev-chargers-from-basements-what-this-means-for-fire-safety-01kqc1gew0k83pr0q72ytgyz33",
		"https://firefibers.com/blogs/news/lithium-ion-battery-fire-erupts-in-ewa-beach-carport-a-wake-up-call-for-homeowners-01kmg8hxf0syvx0w2m0sjyh35x",
		"https://firefibers.com/blogs/news/lithium-ion-battery-fires-a-costly-reminder-from-providence-01kp6a1hj0ree4an774s0f864d",
		"https://firefibers.com/blogs/news/sparta-fire-department-deploys-specialized-ev-blankets-the-uc-now-01kr5qx12d3q3qv4yk6segfs87",
		"https://firefibers.com/blogs/news/tagged/e-bike-fire",
		"https://firefibers.com/blogs/news/tagged/ev-charging",
		"https://firefibers.com/blogs/news/tagged/ev-fire",
		"https://firefibers.com/blogs/news/tagged/ev-fire?page=2",
		"https://firefibers.com/blogs/news/tagged/ev-fire?page=3",
		"https://firefibers.com/blogs/news/tagged/lithium-ion-battery-fire",
		"https://firefibers.com/blogs/news/the-ev-charging-revolution-has-a-surprising-roadblock-the-neighbors-01kmhjcdg0sdpynr0w048j4csz",
		"https://firefibers.com/blogs/news/the-growing-danger-of-ev-battery-fires-what-rhode-islands-experience-tells-us-all-01kpdsm5k0pearysjaqvj615xf",
		"https://firefibers.com/blogs/news/thiells-fire-department-takes-proactive-steps-with-electric-vehicle-fire-safety-training-01kqsatp7014g1a14b96djwyb2",
		"https://firefibers.com/cart",
		"https://firefibers.com/collections/all",
		"https://firefibers.com/collections/frontpage",
		"https://firefibers.com/collections/vendors?q=FireFibers",
		"https://firefibers.com/pages/contact",
		"https://firefibers.com/pages/resources",
		"https://firefibers.com/pages/resources-faq",
		"https://firefibers.com/pages/resources-how-to",
		"https://firefibers.com/pages/specifications",
		"https://firefibers.com/products/battery-fire-bag",
		"https://firefibers.com/products/ev-fire-blanket",
		"https://firefibers.com/search",
	}

	root := NewNode(urls[0])
	for _, url := range urls {
		root.Add(url)
	}
	fmt.Println(util.JSON(root.Children.Values()))
}

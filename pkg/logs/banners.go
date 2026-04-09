package logs

import (
	"strings"
)

var bannerTopLines = []string{
	YellowBoldIntense + ` * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
	YellowBoldIntense + ` *                                                                           * `,
	YellowBoldIntense + ` *` + BlueBoldIntense + `    ██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + BlueBoldIntense + `    ██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + BlueBoldIntense + `    ██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + BlueBoldIntense + `    ██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + BlueBoldIntense + `    ██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + BlueBoldIntense + `    ╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + BlackBoldIntense + `    -------------------------------------------------------------------    ` + YellowBoldIntense + `* `,
}

var managerLines = []string{
	YellowBoldIntense + ` *` + Purple + `               ▄▄   ▄▄  ▄▄▄  ▄▄  ▄▄  ▄▄▄   ▄▄▄▄ ▄▄▄▄▄ ▄▄▄▄                 ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + Purple + `               ██▀▄▀██ ██▀██ ███▄██ ██▀██ ██ ▄▄ ██▄▄  ██▄█▄                ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + Purple + `               ██   ██ ██▀██ ██ ▀██ ██▀██ ▀███▀ ██▄▄▄ ██ ██                ` + YellowBoldIntense + `* `,
}

var workerLines = []string{
	YellowBoldIntense + ` *` + Purple + `                  ▄▄   ▄▄  ▄▄▄  ▄▄▄▄  ▄▄ ▄▄ ▄▄▄▄▄ ▄▄▄▄                     ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + Purple + `                  ██ ▄ ██ ██▀██ ██▄█▄ ██▄█▀ ██▄▄  ██▄█▄                    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + ` *` + Purple + `                   ▀█▀█▀  ▀███▀ ██ ██ ██ ██ ██▄▄▄ ██ ██                    ` + YellowBoldIntense + `* `,
}

var bannerBottomLines = []string{
	YellowBoldIntense + ` *                                                                           * `,
	YellowBoldIntense + ` * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
}

func PrintWorkerBanner() {
	var lines []string
	lines = append(lines, bannerTopLines...)
	lines = append(lines, workerLines...)
	lines = append(lines, bannerBottomLines...)
	println(strings.Join(lines, "\n") + Reset)
}

func PrintManagerBanner() {
	var lines []string
	lines = append(lines, bannerTopLines...)
	lines = append(lines, managerLines...)
	lines = append(lines, bannerBottomLines...)
	println(strings.Join(lines, "\n") + Reset)
}

func Meow() {
	println(strings.Join([]string{
		BlueIntense + `████████████████████████████████████████████████████████████████████████████████`,
		BlueIntense + `████████████████████████████████████████████████████████████████████████████████`,
		Red + `██████████████████` + BlueIntense + `████████████████` + Black + `██████████████████████████████` + BlueIntense + `████████████████`,
		Red + `████████████████████████████████` + Black + `██` + White + `██████████████████████████████` + Black + `██` + BlueIntense + `██████████████`,
		RedIntense + `████` + Red + `██████████████████████████` + Black + `██` + White + `██████` + Purple + `██████████████████████` + White + `██████` + Black + `██` + BlueIntense + `████████████`,
		RedIntense + `██████████████████████████████` + Black + `██` + White + `████` + Purple + `████████████████` + Black + `████` + Purple + `██████` + White + `████` + Black + `██` + BlueIntense + `██` + Black + `████` + BlueIntense + `██████`,
		RedIntense + `██████████████████████████████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `████` + Black + `██` + Purple + `██████` + White + `██` + Black + `████` + White + `████` + Black + `██` + BlueIntense + `████`,
		YellowIntense + `██████████████████` + RedIntense + `████████████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `██████` + Purple + `██████` + White + `██` + Black + `██` + White + `██████` + Black + `██` + BlueIntense + `████`,
		YellowIntense + `██████████████████████` + Black + `██` + YellowIntense + `██████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `██████` + Black + `████████` + White + `████████` + Black + `██` + BlueIntense + `████`,
		YellowIntense + `████████████████████` + Black + `██` + White + `██` + Black + `██` + YellowIntense + `████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `██████████████████████` + Black + `██` + BlueIntense + `████`,
		GreenIntense + `██████████████████` + YellowIntense + `██` + Black + `██` + White + `██` + Black + `████████` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██████████████████████████` + Black + `██` + BlueIntense + `██`,
		GreenIntense + `██████████████████████` + White + `████████` + Black + `██` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██████` + YellowIntense + `██` + White + `██████████` + YellowIntense + `██` + Black + `██` + White + `████` + Black + `██` + BlueIntense + `██`,
		GreenIntense + `██████████████████████` + Black + `████` + White + `████` + Black + `██` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██████` + Black + `██` + White + `██████` + Black + `██` + White + `██` + Black + `████` + White + `████` + Black + `██` + BlueIntense + `██`,
		Blue + `██████████████████` + GreenIntense + `████████` + Black + `██████` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██` + Purple + `████` + White + `████████████████` + Purple + `████` + Black + `██` + BlueIntense + `██`,
		Blue + `██████████████████████████████` + Black + `██` + White + `████` + Purple + `██████████████` + Black + `██` + White + `██████` + Black + `████████████` + White + `████` + Black + `██` + BlueIntense + `████`,
		BlueIntense + `██████████████████` + Blue + `████` + Blue + `██████` + Black + `████` + White + `██████` + Purple + `██████████████` + Black + `██` + White + `██████████████████` + Black + `██` + BlueIntense + `██████`,
		BlueIntense + `██████████████████████████` + Black + `██` + White + `██` + Black + `████` + White + `████████████████████` + Black + `██████████████████` + BlueIntense + `████████`,
		BlueIntense + `████████████████████████` + Black + `██` + White + `██████` + Black + `████████████████████████████████` + White + `██` + Black + `██` + BlueIntense + `████████████`,
		BlueIntense + `████████████████████████` + Black + `██` + White + `████` + Black + `██` + BlueIntense + `██` + Black + `██` + White + `████` + BlueIntense + `████████████` + Black + `██` + White + `████` + Black + `████` + White + `████` + Black + `██` + BlueIntense + `████████████`,
		BlueIntense + `████████████████████████` + Black + `██████` + BlueIntense + `████` + Black + `██████` + BlueIntense + `████████████` + Black + `██████` + BlueIntense + `████` + Black + `██████` + BlueIntense + `████████████`,
		`████████████████████████████████████████████████████████████████████████████████`,
	}, "\n"))
}

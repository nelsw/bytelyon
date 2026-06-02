package email

const template = `
<!doctype html>
<html lang="en">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<body>
<div
        style='background-color:inherit;color:#909090;font-family:Avenir, "Avenir Next LT Pro", Montserrat, Corbel, "URW Gothic", source-sans-pro, sans-serif;font-size:16px;font-weight:400;letter-spacing:0.15008px;line-height:1.5;margin:0;padding:32px 0;min-height:100%;width:100%'
>
    <table
            align="center"
            width="100%"
            style="margin:0 auto;max-width:300px;background-color:#1d1d1d;border-radius:4px;border:1px solid #616161"
            role="presentation"
            cellspacing="0"
            cellpadding="0"
            border="0"
    >
        <tbody>
        <tr style="width:100%">
            <td>
                <div style="padding:24px 24px 24px 24px;text-align:center">
                    <img
                            alt=""
                            src="https://bytelyon-public.s3.amazonaws.com/logo.png"
                            height="128"
                            style="height:128px;outline:none;border:none;text-decoration:none;vertical-align:middle;display:inline-block;max-width:100%"
                    />
                </div>
                <div style="text-align:center;padding:0px 24px 24px 24px">
                    <a
                            href="{{href}}"
                            style="color:#FFFFFF;font-size:16px;font-weight:bold;background-color:#0560ae;border-radius:4px;display:inline-block;padding:12px 20px;text-decoration:none"
                            target="_blank"
                    ><span
                    ><!--[if mso
                      ]><i
                        style="letter-spacing: 20px;mso-font-width:-100%;mso-text-raise:30"
                        hidden
                        >&nbsp;</i
                      ><!
                    [endif]--></span
                    ><span>{{text}}</span
                    ><span
                    ><!--[if mso
                      ]><i
                        style="letter-spacing: 20px;mso-font-width:-100%"
                        hidden
                        >&nbsp;</i
                      ><!
                    [endif]--></span
                    ></a
                    >
                </div>
            </td>
        </tr>
        </tbody>
    </table>
</div>
</body>
</html>`

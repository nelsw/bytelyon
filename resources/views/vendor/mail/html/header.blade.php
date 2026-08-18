@props(['url'])
<tr>
    <td class="header">
        <a href="{{ $url }}" style="display: inline-block;">
            <img src="https://bytelyon-public.s3.amazonaws.com/logo.png" class="logo"
                 alt="{{ config('app.name') }} Logo">
        </a>
    </td>
</tr>

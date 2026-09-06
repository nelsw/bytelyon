<?php

namespace App\Enums;
enum SerpPart: string
{
    case SpoPro = 'sponsored_products';
    case SpoRes = 'sponsored_results';
    case OrgPro = 'organic_products';
    case OrgRes = 'organic_results';
    case SimQry = 'similar_queries';
}

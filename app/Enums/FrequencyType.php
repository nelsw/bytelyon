<?php

namespace App\Enums;

use App\Traits\HasArrays;

enum FrequencyType: string
{
    use HasArrays;

    case Hourly = 'hourly';
    case Daily = 'daily';
    case Weekly = 'weekly';
    case Monthly = 'monthly';
}

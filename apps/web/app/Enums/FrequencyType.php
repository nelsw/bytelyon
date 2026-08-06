<?php

namespace App\Enums;

use App\Traits\HasArrays;
use Carbon\CarbonInterval;

enum FrequencyType: string
{
    use HasArrays;

    case Hourly = 'hourly';
    case Daily = 'daily';
    case Weekly = 'weekly';
    case Monthly = 'monthly';

    public function interval(): CarbonInterval
    {
        return match ($this) {
            FrequencyType::Hourly => CarbonInterval::hour(),
            FrequencyType::Daily => CarbonInterval::day(),
            FrequencyType::Weekly => CarbonInterval::week(),
            FrequencyType::Monthly => CarbonInterval::month(),
        };
    }
}

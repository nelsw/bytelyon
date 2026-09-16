<?php

use App\Providers\AppServiceProvider;
use App\Providers\FortifyServiceProvider;
use App\Providers\MacroServiceProvider;
use App\Providers\TelescopeServiceProvider;

return [
    AppServiceProvider::class,
    FortifyServiceProvider::class,
    MacroServiceProvider::class,
    TelescopeServiceProvider::class,
];

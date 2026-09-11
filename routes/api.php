<?php

use App\Http\Controllers\ApiController;
use Illuminate\Support\Facades\Route;

Route::middleware(['auth:sanctum', 'abilities:worker'])->name('api.')->controller(ApiController::class)->group(function () {});

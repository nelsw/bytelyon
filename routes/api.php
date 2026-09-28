<?php

use Illuminate\Support\Facades\Route;

Route::middleware(['auth:sanctum', 'abilities:worker'])->name('api.')->group(function () {});

<?php

use App\Http\Controllers\Api\ScrapeJobController;
use Illuminate\Support\Facades\Route;

Route::middleware(['auth:sanctum', 'abilities:worker'])->name('api.')->group(function () {
    Route::prefix('scrape-jobs')->name('scrape-jobs.')->group(function () {
        Route::post('serp/{serp}/complete', [ScrapeJobController::class, 'serp'])->name('serp.complete');
        Route::post('news/{article}/complete', [ScrapeJobController::class, 'news'])->name('news.complete');
        Route::post('sitemap/{bot}/complete', [ScrapeJobController::class, 'sitemap'])->name('sitemap.complete');
    });
});

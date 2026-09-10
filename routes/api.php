<?php

use App\Http\Controllers\ApiController;
use Illuminate\Support\Facades\Route;

Route::middleware(['auth:sanctum', 'abilities:worker'])->name('api.')->controller(ApiController::class)->group(function () {

    Route::prefix('bots')->group(function () {
        Route::get('', 'bots')->name('bots.index');
        Route::prefix('{bot}')->group(function () {
            Route::put('', 'bot')->name('bots.update');
            Route::put('/articles', 'article')->name('articles.upsert');
            Route::put('/searches', 'serp')->name('searches.upsert');
            Route::put('/sitemaps', 'sitemap')->name('sitemaps.upsert');
        });
    });

    Route::put('/searches/{serp}/page', 'serpPage')->name('searches.pages.upsert');
    Route::put('/sitemaps/{sitemap}/page', 'sitemapPage')->name('sitemaps.pages.upsert');

    Route::post('/pages/{page}', 'pageImg')->name('pages.images.upsert');
    Route::post('/searches/{serp}', 'searchImg')->name('searches.images.upsert');
});

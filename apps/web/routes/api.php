<?php

use App\Http\Controllers\ApiController;
use App\Http\Middleware\CheckApiKey;

Route::middleware(CheckApiKey::class)->name('api.')->controller(ApiController::class)->group(function () {

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
});

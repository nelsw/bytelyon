<?php

use App\Http\Controllers\ArticleController;
use App\Http\Controllers\BotController;
use App\Http\Controllers\DashboardController;
use App\Http\Controllers\NewsController;
use App\Http\Controllers\SerpController;
use App\Http\Controllers\SitemapController;
use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::get('/', fn () => Inertia::render('Welcome'))->name('home');

Route::middleware(['auth', 'verified'])->group(function () {
    Route::get('dashboard', [DashboardController::class, 'index'])->name('dashboard');

    Route::get('news', [NewsController::class, 'index'])->name('news.index');

    Route::prefix('bots/{bot}/articles')
        ->name('articles.')
        ->controller(ArticleController::class)
        ->group(function () {
            Route::get('/', 'index')->name('index');
            Route::get('/{article}', 'show')->name('show');
            Route::get('/{article}/edit', 'edit')->name('edit');
            Route::put('/{article}', 'update')->name('update');
            Route::post('/{article}/assist', 'assist')->name('assist');
            Route::post('/{article}/publish', 'publish')->name('publish');
            Route::delete('/{article}', 'destroy')->name('destroy');
        });

    Route::prefix('sitemaps')
        ->name('sitemaps.')
        ->controller(SitemapController::class)
        ->group(function () {
            Route::get('/', 'index')->name('index');
            Route::get('/{sitemap}', 'show')->name('show');
            Route::delete('/{sitemap}', 'destroy')->name('destroy');
        });

    Route::prefix('serps')
        ->name('serps.')
        ->controller(SerpController::class)
        ->group(function () {
            Route::get('/', 'index')->name('index');
            Route::get('/{serp}', 'show')->name('show');
            Route::delete('/{serp}', 'destroy')->name('destroy');
        });

    Route::softDeletableResources([
        'bots' => BotController::class,
    ]);
});

require __DIR__.'/settings.php';

<?php

use App\Http\Controllers\Settings\ApiTokenController;
use App\Http\Controllers\Settings\IntegrationsController;
use App\Http\Controllers\Settings\ProfileController;
use App\Http\Controllers\Settings\ProxyController;
use App\Http\Controllers\Settings\SecurityController;
use Illuminate\Auth\Middleware\RequirePassword;
use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::middleware(['auth'])->group(function () {
    Route::redirect('settings', '/settings/profile');

    Route::get('settings/profile', [ProfileController::class, 'edit'])->name('profile.edit');
    Route::patch('settings/profile', [ProfileController::class, 'update'])->name('profile.update');

    Route::get('settings/integrations', [IntegrationsController::class, 'edit'])->name('integrations.edit');
    Route::put('settings/integrations/anthropic', [IntegrationsController::class, 'updateAnthropic'])->name('integrations.anthropic.update');
    Route::put('settings/integrations/shopify', [IntegrationsController::class, 'updateShopify'])->name('integrations.shopify.update');
});

Route::middleware(['auth', 'verified'])->group(function () {
    Route::delete('settings/profile', [ProfileController::class, 'destroy'])->name('profile.destroy');

    Route::get('settings/security', [SecurityController::class, 'edit'])
        ->middleware(RequirePassword::class)
        ->name('security.edit');

    Route::put('settings/password', [SecurityController::class, 'update'])
        ->middleware('throttle:6,1')
        ->name('user-password.update');

    Route::get('settings/appearance', fn () => Inertia::render('settings/Appearance'))->name('appearance.edit');

    Route::get('settings/api-tokens', [ApiTokenController::class, 'edit'])->name('api-tokens.edit');
    Route::post('settings/api-tokens', [ApiTokenController::class, 'store'])->name('api-tokens.store');
    Route::delete('settings/api-tokens/{token}', [ApiTokenController::class, 'destroy'])
        ->whereNumber('token')
        ->name('api-tokens.destroy');

    Route::get('settings/proxies', [ProxyController::class, 'edit'])->name('proxies.edit');
    Route::post('settings/proxies', [ProxyController::class, 'store'])->name('proxies.store');
    Route::delete('settings/proxies/{proxy}', [ProxyController::class, 'destroy'])->name('proxies.destroy');
});

Route::get('.well-known/passkey-endpoints', function () {
    return response()->json([
        'enroll' => route('security.edit'),
        'manage' => route('security.edit'),
    ]);
})->name('well-known.passkeys');

<?php

use App\Models\Bot;
use App\Models\Proxy;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bot_proxy', function (Blueprint $table) {
            $table->foreignIdFor(Bot::class)->constrained()->cascadeOnDelete();
            $table->foreignIdFor(Proxy::class)->constrained()->cascadeOnDelete();
            $table->primary(['bot_id', 'proxy_id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bot_proxy');
    }
};

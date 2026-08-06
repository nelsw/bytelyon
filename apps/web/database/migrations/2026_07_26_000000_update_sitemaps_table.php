<?php

use App\Models\Bot;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('sitemaps', function (Blueprint $table) {
            $table->dropUnique(['domain']);
            $table->foreignIdFor(Bot::class)->constrained()->cascadeOnDelete();
            $table->unique(['bot_id', 'domain']);
        });
    }

    public function down(): void {}
};

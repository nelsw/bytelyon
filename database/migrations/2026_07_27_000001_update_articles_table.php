<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('articles', function (Blueprint $table) {
            $table->string('img_alt')->nullable()->change();
            $table->string('img_url')->nullable()->change();
            $table->string('source')->nullable()->change();
            $table->string('keywords')->nullable()->change();
            $table->string('description')->nullable()->change();
            $table->text('body')->nullable()->change();
            $table->unique(['link', 'bot_id']);
        });
    }

    public function down(): void {}
};

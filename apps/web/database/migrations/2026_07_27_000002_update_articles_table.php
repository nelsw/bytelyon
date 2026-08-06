<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('articles', function (Blueprint $table) {
            $table->dropUnique(['link', 'bot_id']);
            $table->dropColumn('link');

            $table->string('url');
            $table->unique(['url', 'bot_id']);
        });
    }

    public function down(): void
    {
        Schema::table('articles', function (Blueprint $table) {
            $table->dropUnique(['url', 'bot_id']);
            $table->dropColumn('url');

            $table->string('link');
            $table->unique(['link', 'bot_id']);
        });
    }
};

<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('bots', function (Blueprint $table) {
            $table->dropColumn('played_at');
            $table->dropColumn('play_result');
            $table->dropColumn('last_run_result');
        });
    }

    public function down(): void
    {
        Schema::table('bots', function (Blueprint $table) {
            $table->timestamp('played_at')->nullable();
            $table->string('play_result')->nullable();
            $table->string('last_run_result')->nullable();
        });
    }
};

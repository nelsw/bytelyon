<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('articles', function (Blueprint $table) {
            $table->string('url', 2048)->change();
            $table->string('keywords', 1024)->nullable()->change();
            $table->string('img_alt', 1024)->nullable()->change();
            $table->text('description')->nullable()->change();
        });
    }

    public function down(): void {}
};

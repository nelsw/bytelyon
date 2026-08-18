<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('serps', function (Blueprint $table) {
            $table->string('screenshot_key')->nullable()->change();
            $table->string('content_key')->nullable()->change();
            $table->jsonb('data')->nullable()->change();
        });
    }

    public function down(): void {}
};

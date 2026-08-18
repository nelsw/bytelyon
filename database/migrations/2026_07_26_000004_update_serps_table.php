<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('serps', function (Blueprint $table) {
            $table->dropColumn('page_ids');
            $table->string('content_key');
        });
    }

    public function down(): void
    {
        Schema::table('serps', function (Blueprint $table) {
            $table->string('page_ids')->nullable();
            $table->dropColumn('content_key');
        });
    }
};

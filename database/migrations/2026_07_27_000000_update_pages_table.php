<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('pages', function (Blueprint $table) {
            $table->string('screenshot_key')->nullable()->change();
            $table->json('meta')->nullable()->change();
            $table->unique(['pageable_type', 'pageable_id', 'url']);
        });
    }

    public function down(): void {}
};

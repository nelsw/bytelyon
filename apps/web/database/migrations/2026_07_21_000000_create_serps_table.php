<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('serps', function (Blueprint $table) {
            $table->id();
            $table->string('query');
            $table->string('screenshot_key');
            $table->jsonb('data');
            $table->jsonb('page_ids')->nullable();
            $table->timestamps();
            $table->softDeletes();
            $table->foreignId('bot_id')->constrained()->cascadeOnDelete();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('serps');
    }
};

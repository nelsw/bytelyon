<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('articles', function (Blueprint $table) {
            $table->id();
            $table->string('title');
            $table->string('link');
            $table->timestamp('published_at');
            $table->string('img_alt');
            $table->string('img_url');
            $table->string('source');
            $table->string('keywords');
            $table->string('description');
            $table->text('body');
            $table->timestamps();
            $table->softDeletes();
            $table->foreignId('bot_id')->constrained()->cascadeOnDelete();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('articles');
    }
};

<?php

use App\Models\Bot;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('news', function (Blueprint $table) {
            $table->id();
            $table->string('description');
            $table->string('link');
            $table->timestamp('published_at');
            $table->string('query');
            $table->string('source');
            $table->string('title');
            $table->enum('type', ['google_news', 'bing_news']);
            $table->timestamp('created_at')->useCurrent();
            $table->foreignIdFor(Bot::class)->constrained()->cascadeOnDelete();
            $table->unique(['bot_id', 'type', 'query', 'link']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('news');
    }
};

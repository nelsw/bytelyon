<?php

use App\Enums\BotType;
use App\Enums\FrequencyType;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bots', function (Blueprint $table) {
            $table->id();
            $table->string('blacklist')->nullable();
            $table->boolean('enabled');
            $table->boolean('headless');
            $table->enum('frequency', FrequencyType::values());
            $table->string('query');
            $table->enum('type', BotType::values());
            $table->timestamp('last_run_at')->nullable();
            $table->timestamps();
            $table->softDeletes();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->unique(['user_id', 'query', 'type', 'deleted_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bots');
    }
};

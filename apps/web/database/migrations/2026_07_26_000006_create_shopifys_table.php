<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('shopifys', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('client_id');
            $table->string('client_secret');
            $table->string('default_author_name')->nullable();
            $table->string('default_blog_id')->nullable();
            $table->string('store');
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('shopifys');
    }
};

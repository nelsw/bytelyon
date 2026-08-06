<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('bots', function (Blueprint $table) {
            $table->dropUnique(['user_id', 'query', 'type', 'deleted_at']);
            $query = <<<'SQL'
            CREATE UNIQUE INDEX bots_user_id_query_type ON bots (user_id, query, type) WHERE deleted_at IS NULL
SQL;
            DB::statement($query);
        });
    }

    public function down(): void
    {
        Schema::table('users', function (Blueprint $table) {
            $table->dropUnique('bots_user_id_query_type');
            $table->unique(['user_id', 'query', 'type', 'deleted_at']);
        });
    }
};

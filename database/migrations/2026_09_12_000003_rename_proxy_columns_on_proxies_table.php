<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('proxies', function (Blueprint $table) {
            $table->renameColumn('username', 'user');
            $table->renameColumn('password', 'pass');
            $table->renameColumn('server', 'host');
            $table->renameColumn('protocol', 'scheme');
        });
    }

    public function down(): void
    {
        Schema::table('proxies', function (Blueprint $table) {
            $table->renameColumn('user', 'username');
            $table->renameColumn('pass', 'password');
            $table->renameColumn('host', 'server');
            $table->renameColumn('scheme', 'protocol');
        });
    }
};

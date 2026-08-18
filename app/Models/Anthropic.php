<?php

namespace App\Models;

use App\Traits\HasUser;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Model;

#[Fillable('api_key', 'default_model')]
class Anthropic extends Model
{
    use HasUser;
}

<?php

namespace App\Models;

use App\Traits\HasUser;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\Table;
use Illuminate\Database\Eloquent\Model;

#[Fillable('client_id', 'client_secret', 'default_author_name', 'default_blog_id', 'store')]
#[Table('shopifys')]
class Shopify extends Model
{
    use HasUser;
}

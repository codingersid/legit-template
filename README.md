# Legit Template Engine

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version">
</p>

**Legit Template** adalah template engine untuk Go yang terinspirasi dari Laravel Blade. Template engine ini dirancang khusus untuk [Legit Framework](https://github.com/codingersid/legit), namun dapat digunakan secara mandiri atau dengan framework Go lainnya seperti Fiber, Echo, dan Gin.

## Fitur Utama

- **Sintaks Blade-like** - Familiar bagi developer Laravel
- **Template Inheritance** - `@extends`, `@section`, `@yield`, `@include`
- **Kontrol Struktur** - `@if`, `@foreach`, `@for`, `@switch`, dan lainnya
- **Variabel $loop** - Akses informasi loop seperti `$loop.index`, `$loop.first`, `$loop.last`
- **Komponen & Slot** - Buat komponen reusable dengan `@component` dan `@slot`
- **Stack** - Kelola scripts dan styles dengan `@push`, `@prepend`, `@stack`
- **80+ Fungsi Bawaan** - Manipulasi string, array, tanggal, angka, dan lainnya
- **Caching** - Template caching untuk performa optimal
- **Integrasi Fiber** - Adapter khusus untuk Fiber framework

## Instalasi

```bash
go get github.com/codingersid/legit-template
```

## Penggunaan Dasar

### Standalone

```go
package main

import (
    "os"
    legit "github.com/codingersid/legit-template"
)

func main() {
    // Buat engine baru
    engine := legit.New("./views")

    // Render template
    data := map[string]interface{}{
        "title": "Selamat Datang",
        "name":  "John Doe",
    }

    engine.Render(os.Stdout, "pages.home", data)
}
```

### Dengan Fiber Framework

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/codingersid/legit-template/fiber"
)

func main() {
    // Buat template engine
    engine := fiber.New("./views", ".legit")

    // Development mode (disable cache)
    engine.Reload(true)

    // Buat Fiber app
    app := fiber.New(fiber.Config{
        Views: engine,
    })

    app.Get("/", func(c *fiber.Ctx) error {
        return c.Render("pages/home", fiber.Map{
            "title": "Beranda",
            "user":  "John",
        })
    })

    app.Listen(":3000")
}
```

## Sintaks Template

### Output

```blade
{{-- Komentar (tidak dirender) --}}

{{ $variable }}          {{-- Output escaped (aman dari XSS) --}}
{!! $html !!}            {{-- Output raw/unescaped --}}
@{{ $literal }}          {{-- Output literal {{ }} --}}
```

### Template Inheritance

**layouts/app.legit:**
```blade
<!DOCTYPE html>
<html>
<head>
    <title>@yield('title', 'Default Title')</title>
    @stack('styles')
</head>
<body>
    @include('partials.navbar')

    <main>
        @yield('content')
    </main>

    @include('partials.footer')
    @stack('scripts')
</body>
</html>
```

**pages/home.legit:**
```blade
@extends('layouts.app')

@section('title', 'Beranda')

@section('content')
<div class="container">
    <h1>Selamat Datang, {{ $user }}</h1>
</div>
@endsection

@push('scripts')
<script>console.log('Home loaded');</script>
@endpush
```

### Kondisional

```blade
@if($user.isAdmin)
    <span>Admin</span>
@elseif($user.isModerator)
    <span>Moderator</span>
@else
    <span>Member</span>
@endif

@unless($user.isGuest)
    <p>Selamat datang kembali!</p>
@endunless

@isset($title)
    <h1>{{ $title }}</h1>
@endisset

@empty($items)
    <p>Tidak ada item</p>
@endempty
```

### Switch Case

```blade
@switch($status)
    @case('pending')
        <span class="badge yellow">Menunggu</span>
        @break
    @case('approved')
        <span class="badge green">Disetujui</span>
        @break
    @case('rejected')
        <span class="badge red">Ditolak</span>
        @break
    @default
        <span class="badge gray">Unknown</span>
@endswitch
```

### Perulangan

```blade
{{-- For Loop --}}
@for($i = 0; $i < 10; $i++)
    <p>Iterasi ke-{{ $i }}</p>
@endfor

{{-- Foreach Loop --}}
@foreach($users as $user)
    <div class="user">
        <p>{{ $loop.iteration }}. {{ $user.name }}</p>

        @if($loop.first)
            <span>Pertama!</span>
        @endif

        @if($loop.last)
            <span>Terakhir!</span>
        @endif
    </div>
@endforeach

{{-- Foreach dengan key --}}
@foreach($items as $key => $value)
    <p>{{ $key }}: {{ $value }}</p>
@endforeach

{{-- Forelse (dengan fallback jika kosong) --}}
@forelse($products as $product)
    <div class="product">{{ $product.name }}</div>
@empty
    <p>Tidak ada produk tersedia.</p>
@endforelse

{{-- While Loop --}}
@while($condition)
    <p>Looping...</p>
@endwhile
```

### Variabel $loop

Variabel `$loop` tersedia di dalam semua perulangan:

| Properti | Deskripsi |
|----------|-----------|
| `$loop.index` | Index iterasi saat ini (dimulai dari 0) |
| `$loop.iteration` | Nomor iterasi saat ini (dimulai dari 1) |
| `$loop.remaining` | Sisa iterasi |
| `$loop.count` | Total jumlah item |
| `$loop.first` | Apakah ini iterasi pertama |
| `$loop.last` | Apakah ini iterasi terakhir |
| `$loop.even` | Apakah ini iterasi genap |
| `$loop.odd` | Apakah ini iterasi ganjil |
| `$loop.depth` | Kedalaman nesting loop |
| `$loop.parent` | Variabel $loop parent (untuk nested loop) |

### Include

```blade
{{-- Include sederhana --}}
@include('partials.header')

{{-- Include dengan data tambahan --}}
@include('partials.user-card', ['user' => $currentUser])

{{-- Include jika ada --}}
@includeIf('partials.optional')

{{-- Include dengan kondisi --}}
@includeWhen($showSidebar, 'partials.sidebar')
@includeUnless($isGuest, 'partials.user-menu')

{{-- Include pertama yang ada --}}
@includeFirst(['custom.header', 'default.header'])

{{-- Include untuk setiap item --}}
@each('partials.item', $items, 'item', 'partials.no-items')
```

### Komponen & Slot

**components/alert.legit:**
```blade
<div class="alert alert-{{ $type ?? 'info' }}">
    @if(isset($slots.title))
        <h4>{{ $slots.title }}</h4>
    @endif

    <div class="alert-body">
        {{ $slot }}
    </div>
</div>
```

**Penggunaan:**
```blade
@component('components.alert', ['type' => 'success'])
    @slot('title')
        Berhasil!
    @endslot

    Data berhasil disimpan.
@endcomponent
```

### Stack (Scripts & Styles)

**Layout:**
```blade
<head>
    {{-- CSS stack --}}
    @stack('styles')
</head>
<body>
    @yield('content')

    {{-- JS stack --}}
    @stack('scripts')
</body>
```

**Page:**
```blade
@push('styles')
<link rel="stylesheet" href="/css/page.css">
@endpush

@push('scripts')
<script src="/js/page.js"></script>
@endpush

{{-- Push hanya sekali (deduplicate) --}}
@pushOnce('scripts')
<script src="/js/library.js"></script>
@endPushOnce

{{-- Prepend (tambah di awal) --}}
@prepend('scripts')
<script>var config = {};</script>
@endprepend
```

### Autentikasi

```blade
@auth
    <p>Selamat datang, {{ $auth.name }}!</p>
    <a href="/logout">Logout</a>
@endauth

@guest
    <a href="/login">Login</a>
    <a href="/register">Daftar</a>
@endguest

{{-- Dengan guard tertentu --}}
@auth('admin')
    <a href="/admin">Dashboard Admin</a>
@endauth
```

### Environment

```blade
@env('local')
    <p>Mode Development</p>
@endenv

@env(['local', 'staging'])
    <p>Debug Info tersedia</p>
@endenv

@production
    {{-- Hanya di production --}}
    <script src="/js/analytics.js"></script>
@endproduction
```

### Form Helpers

```blade
<form method="POST" action="/users">
    {{-- CSRF Token --}}
    @csrf

    {{-- Method Spoofing --}}
    @method('PUT')

    <input type="text" name="name" value="{{ old 'name' }}">

    {{-- Tampilkan error validasi --}}
    @error('name')
        <span class="error">{{ $message }}</span>
    @enderror

    <button type="submit">Simpan</button>
</form>
```

### Attribute Helpers

```blade
<input type="checkbox" @checked($isActive)>
<option @selected($option eq $selected)>{{ $option }}</option>
<input type="text" @disabled($isReadOnly)>
<input type="text" @readonly($isLocked)>
<input type="text" @required($isRequired)>
```

### Fungsi Utilitas

```blade
{{-- JSON output --}}
<script>
    var data = @json($data);
</script>

{{-- Verbatim (tidak diparse) --}}
@verbatim
    <div id="app">
        {{ message }}  {{-- Ini akan dirender oleh Vue.js --}}
    </div>
@endverbatim

{{-- Render sekali --}}
@once
    <script src="/js/shared.js"></script>
@endonce
```

## Fungsi Bawaan

### String

| Fungsi | Deskripsi | Contoh |
|--------|-----------|--------|
| `upper` | Huruf besar | `{{ upper $name }}` |
| `lower` | Huruf kecil | `{{ lower $name }}` |
| `title` | Title Case | `{{ title $name }}` |
| `trim` | Hapus whitespace | `{{ trim $text }}` |
| `substr` | Substring | `{{ substr $text 0 10 }}` |
| `length` | Panjang string | `{{ length $text }}` |
| `replace` | Ganti string | `{{ replace $text "old" "new" }}` |
| `contains` | Cek substring | `{{ if contains $text "kata" }}` |
| `split` | Pecah string | `{{ split $text "," }}` |
| `join` | Gabung array | `{{ join $arr ", " }}` |
| `slug` | Buat slug | `{{ slug $title }}` |
| `limit` | Potong dengan ... | `{{ limit $text 100 }}` |
| `nl2br` | Newline to BR | `{!! nl2br $text !!}` |

### Array

| Fungsi | Deskripsi | Contoh |
|--------|-----------|--------|
| `first` | Elemen pertama | `{{ first $arr }}` |
| `last` | Elemen terakhir | `{{ last $arr }}` |
| `reverse` | Balik array | `{{ reverse $arr }}` |
| `sortAsc` | Urutkan ascending | `{{ sortAsc $arr }}` |
| `sortDesc` | Urutkan descending | `{{ sortDesc $arr }}` |
| `unique` | Hapus duplikat | `{{ unique $arr }}` |
| `pluck` | Ambil kolom | `{{ pluck $users "name" }}` |
| `where` | Filter array | `{{ where $users "active" true }}` |
| `groupBy` | Kelompokkan | `{{ groupBy $items "category" }}` |
| `chunk` | Bagi array | `{{ chunk $items 3 }}` |
| `merge` | Gabung map | `{{ merge $map1 $map2 }}` |

### Angka

| Fungsi | Deskripsi | Contoh |
|--------|-----------|--------|
| `add` | Tambah | `{{ add $a $b }}` |
| `sub` | Kurang | `{{ sub $a $b }}` |
| `mul` | Kali | `{{ mul $a $b }}` |
| `div` | Bagi | `{{ div $a $b }}` |
| `mod` | Modulo | `{{ mod $a $b }}` |
| `round` | Pembulatan | `{{ round $num 2 }}` |
| `floor` | Bulatkan ke bawah | `{{ floor $num }}` |
| `ceil` | Bulatkan ke atas | `{{ ceil $num }}` |
| `currency` | Format mata uang | `{{ currency $price "Rp" }}` |
| `number` | Format angka | `{{ number $num 2 }}` |
| `percent` | Format persen | `{{ percent $ratio 1 }}` |

### Tanggal

| Fungsi | Deskripsi | Contoh |
|--------|-----------|--------|
| `date` | Format tanggal | `{{ date "d M Y" $time }}` |
| `now` | Waktu sekarang | `{{ now }}` |
| `ago` | Waktu relatif | `{{ ago $time }}` |
| `addDate` | Tambah tanggal | `{{ addDate $time 0 1 0 }}` |
| `timestamp` | Unix timestamp | `{{ timestamp $time }}` |

### Utilitas

| Fungsi | Deskripsi | Contoh |
|--------|-----------|--------|
| `default` | Nilai default | `{{ default $name "Guest" }}` |
| `isset` | Cek ada | `{{ if isset $var }}` |
| `empty` | Cek kosong | `{{ if empty $arr }}` |
| `json` | Encode JSON | `{{ json $data }}` |
| `dump` | Debug dump | `{{ dump $var }}` |
| `coalesce` | Nilai pertama | `{{ coalesce $a $b $c }}` |
| `ternary` | If-else inline | `{{ ternary $cond "ya" "tidak" }}` |

## Konfigurasi

### Opsi Engine

```go
engine := legit.New("./views",
    // Ekstensi file (default: .legit)
    legit.WithExtension(".legit"),

    // Mode development (disable cache)
    legit.WithDevelopment(true),

    // Tambah fungsi kustom
    legit.WithFunctions(template.FuncMap{
        "rupiah": formatRupiah,
        "gravatar": getGravatar,
    }),
)
```

### Opsi Fiber Adapter

```go
engine := fiber.New("./views", ".legit")

// Set default layout
engine.Layout("layouts/main")

// Enable reload mode (development)
engine.Reload(true)

// Enable debug mode
engine.Debug(true)

// Tambah fungsi kustom
engine.AddFunc("custom", myFunc)
engine.AddFuncMap(myFuncs)
```

## Struktur Direktori yang Disarankan

```
views/
├── layouts/
│   ├── app.legit           # Layout utama
│   ├── admin.legit         # Layout admin
│   └── auth.legit          # Layout auth
├── pages/
│   ├── home.legit          # Halaman beranda
│   ├── about.legit         # Halaman about
│   └── contact.legit       # Halaman kontak
├── partials/
│   ├── navbar.legit        # Navigasi
│   ├── footer.legit        # Footer
│   └── sidebar.legit       # Sidebar
├── components/
│   ├── alert.legit         # Komponen alert
│   ├── card.legit          # Komponen card
│   ├── modal.legit         # Komponen modal
│   └── pagination.legit    # Komponen pagination
└── errors/
    ├── 404.legit           # Halaman 404
    ├── 500.legit           # Halaman 500
    └── error.legit         # Halaman error umum
```

## CLI Commands (Legit Framework)

Jika menggunakan Legit Framework, tersedia CLI commands:

```bash
# Buat view baru
legit view:create pages/dashboard
legit view:create components/button
legit view:create layouts/admin

# List semua view
legit view:list

# Cache semua view (production)
legit view:cache

# Clear cache view
legit view:clear
```

## Contoh Lengkap

### E-Commerce Product List

```blade
@extends('layouts.app')

@section('title', 'Daftar Produk')

@section('content')
<div class="container">
    <h1>Produk Kami</h1>

    {{-- Filter --}}
    <div class="filters">
        @foreach($categories as $category)
            <a href="?category={{ $category.slug }}"
               class="@if($current_category eq $category.slug)active@endif">
                {{ $category.name }}
            </a>
        @endforeach
    </div>

    {{-- Product Grid --}}
    <div class="product-grid">
        @forelse($products as $product)
            <div class="product-card">
                <img src="{{ $product.image }}" alt="{{ $product.name }}">
                <h3>{{ $product.name }}</h3>
                <p class="price">{{ currency $product.price "Rp" }}</p>

                @if($product.stock gt 0)
                    <span class="badge green">Tersedia</span>
                @else
                    <span class="badge red">Habis</span>
                @endif

                @auth
                    <button onclick="addToCart({{ $product.id }})">
                        Tambah ke Keranjang
                    </button>
                @else
                    <a href="/login">Login untuk membeli</a>
                @endauth
            </div>
        @empty
            <div class="empty-state">
                <p>Tidak ada produk ditemukan.</p>
            </div>
        @endforelse
    </div>

    {{-- Pagination --}}
    @include('components.pagination', ['paginator' => $products])
</div>
@endsection

@push('scripts')
<script>
    function addToCart(productId) {
        // AJAX call
    }
</script>
@endpush
```

## Lisensi

MIT License - Silakan gunakan untuk proyek apapun.

## Kontribusi

Kontribusi sangat diterima! Silakan buat pull request atau issue di GitHub.

## Kredit

- Terinspirasi oleh [Laravel Blade](https://laravel.com/docs/blade)
- Dikembangkan untuk [Legit Framework](https://github.com/codingersid/legit)
- Dibuat oleh [Codingersid](https://github.com/codingersid)

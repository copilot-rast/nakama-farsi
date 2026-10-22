param(
    [Parameter(Mandatory = $true)][string]$SourceRoot,
    [Parameter(Mandatory = $true)][string]$OutputPath
)

$ErrorActionPreference = 'Stop'
$backend = Join-Path $SourceRoot 'nest-backend/back-end'
$next = Join-Path $SourceRoot 'nextjs'
if (-not (Test-Path (Join-Path $backend 'apps/api/src/app.module.ts')) -or -not (Test-Path (Join-Path $next 'src/Modules'))) {
    throw 'SourceRoot must point to the jazebe directory containing nest-backend and nextjs.'
}

$lines = [System.Collections.Generic.List[string]]::new()
$lines.Add('# Jazebe feature inventory')
$lines.Add('')
$lines.Add('Generated from source controller registrations, worker providers, and Next.js feature modules. These rows identify implementation entry points; they do not assert feature parity.')
$lines.Add('')
$lines.Add('## Backend HTTP controllers')
$lines.Add('')
$lines.Add('| ID | Capability / route group | Source implementation | Route handlers | Destination support | Strategy and destination | Schema / API / security | Tests and evidence | Status |')
$lines.Add('|---|---|---|---:|---|---|---|---|---|')
$controllers = Get-ChildItem (Join-Path $backend 'apps/api/src/modules') -Filter '*.controller.ts' -Recurse | Sort-Object FullName
$index = 0
foreach ($file in $controllers) {
    $index++
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $relative = $file.FullName.Substring($SourceRoot.Length + 1).Replace('\', '/')
    $controller = [regex]::Match($content, '@Controller\(\s*[\x27\x22]([^\x27\x22]+)')
    $route = if ($controller.Success) { $controller.Groups[1].Value } else { '(dynamic controller path)' }
    $handlers = [regex]::Matches($content, '@(Get|Post|Put|Patch|Delete)\s*\(').Count
    $name = $file.BaseName.Replace('.controller', '').Replace('-', ' ')
    $native = $relative -match '/jazebeh/(brand|planet)/'
    $strategy = if ($native) { 'Partial native Go RPC in integrations/jazebe/native/module.go' } else { 'Native Nakama adapter pending' }
    $state = if ($native) { 'Partial' } else { 'Not started' }
    $lines.Add("| JZ-API-$($index.ToString('000')) | $name ($route) | $relative | $handlers | Domain overlap unverified | $strategy | Nakama user IDs and migration required; source HTTP parity pending | Native package compiles; behavior parity unverified | $state |")
}
$lines.Add('')
$lines.Add('## Worker jobs and schedulers')
$lines.Add('')
$lines.Add('| ID | Capability | Source implementation | Source dependencies | Destination implementation | Evidence | Status |')
$lines.Add('|---|---|---|---|---|---|---|')
$workers = Get-ChildItem (Join-Path $backend 'apps/worker/src') -File -Include '*.worker.ts','*.scheduler.ts','*reconcile.service.ts','sweeper.service.ts','job-runner.service.ts' -Recurse | Sort-Object Name
$index = 0
foreach ($file in $workers) {
    $index++
    $relative = $file.FullName.Substring($SourceRoot.Length + 1).Replace('\', '/')
    $lines.Add("| JZ-JOB-$($index.ToString('000')) | $($file.BaseName.Replace('-', ' ')) | $relative | PostgreSQL, Redis, worker lifecycle | Native Nakama lifecycle adapter pending | None | Not started |")
}
$lines.Add('')
$lines.Add('## Next.js user experience modules')
$lines.Add('')
$lines.Add('| ID | Capability | Source implementation | Destination implementation | Evidence | Status |')
$lines.Add('|---|---|---|---|---|---|')
$modules = Get-ChildItem (Join-Path $next 'src/Modules') -Directory | Sort-Object Name
$index = 0
foreach ($module in $modules) {
    $index++
    $relative = $module.FullName.Substring($SourceRoot.Length + 1).Replace('\', '/')
    $lines.Add("| JZ-UI-$($index.ToString('000')) | $($module.Name) | $relative | Native Nakama-compatible UI pending | None | Not started |")
}
$lines.Add('')
$lines.Add('## Administrative frontend modules')
$lines.Add('')
$lines.Add('| ID | Capability | Source implementation | Destination implementation | Evidence | Status |')
$lines.Add('|---|---|---|---|---|---|')
$adminModules = Get-ChildItem (Join-Path $backend '../front-end/src/modules') -Directory | Sort-Object Name
$index = 0
foreach ($module in $adminModules) {
    $index++
    $relative = $module.FullName.Substring($SourceRoot.Length + 1).Replace('\', '/')
    $lines.Add("| JZ-ADMIN-$($index.ToString('000')) | $($module.Name) | $relative | Native Nakama Console extension pending | None | Not started |")
}
$lines.Add('')
$lines.Add("Inventory totals: $($controllers.Count) backend controllers, $($workers.Count) worker entry points, $($modules.Count) Next.js feature modules, $($adminModules.Count) administrative frontend modules. Counted HTTP handlers are decorator occurrences, not individually verified routes.")
$lines.Add('')
$lines.Add('Cross-cutting source dependencies: PostgreSQL/TypeORM migrations, Redis/BullMQ, S3-compatible object storage, SMS provider, JWT/OTP identity, tenancy and RBAC, Socket.IO realtime, external brand API keys and HMAC, webhook delivery, metrics, localization, and separate API and worker lifecycles. These need separate behavioral audits before any row can be marked complete.')
$directory = Split-Path -Parent $OutputPath
New-Item -ItemType Directory -Force -Path $directory | Out-Null
[System.IO.File]::WriteAllLines($OutputPath, $lines, [System.Text.UTF8Encoding]::new($false))

$routeLines = [System.Collections.Generic.List[string]]::new()
$routeLines.Add('# Jazebe HTTP route declaration inventory')
$routeLines.Add('')
$routeLines.Add('This lists controller decorators from the read-only Jazebe NestJS source. A decorator is an entry point, not proof that its operation is integrated with Nakama identity or MCP. Global API prefix and controller path are applied at runtime. Empty method paths use the controller root.')
$routeLines.Add('')
$routeLines.Add('| ID | Method | Controller path | Method path | Source | Status |')
$routeLines.Add('|---|---|---|---|---|---|')
$routeIndex = 0
foreach ($file in $controllers) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $relative = $file.FullName.Substring($SourceRoot.Length + 1).Replace('\', '/')
    $controller = [regex]::Match($content, '@Controller\(\s*[\x27\x22]([^\x27\x22]+)')
    $rootRoute = if ($controller.Success) { $controller.Groups[1].Value } else { '(dynamic)' }
    $decorators = [regex]::Matches($content, '@(Get|Post|Put|Patch|Delete)\s*\(\s*(?:[\x27\x22]([^\x27\x22]*)[\x27\x22])?')
    foreach ($decorator in $decorators) {
        $routeIndex++
        $method = $decorator.Groups[1].Value.ToUpperInvariant()
        $path = $decorator.Groups[2].Value
        $routeLines.Add("| JZ-ROUTE-$($routeIndex.ToString('000')) | $method | $rootRoute | $path | $relative | Native parity pending |")
    }
}
$routeLines.Add('')
$routeLines.Add("Total: $routeIndex route decorators. Guards, DTO validation, service calls, and tenant policies must be traced in source before any adapter is declared complete.")
[System.IO.File]::WriteAllLines((Join-Path $directory 'jazebe-http-routes.md'), $routeLines, [System.Text.UTF8Encoding]::new($false))
Write-Output "Wrote $OutputPath"

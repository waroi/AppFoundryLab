import { DEFAULT_LOCALE, type Locale } from "@/lib/ui/preferences";

export type { Locale } from "@/lib/ui/preferences";

export type PageTitleKey = "home" | "test";

const LOCALE_TAGS: Record<Locale, string> = {
  en: "en-US",
  tr: "tr-TR",
};

const copy = {
  en: {
    common: {
      error: "Error",
      none: "none",
      notAvailable: "n/a",
      unknown: "Unknown",
      yes: "yes",
      no: "no",
      on: "on",
      off: "off",
      enabled: "enabled",
      disabled: "disabled",
      required: "required",
      optional: "optional",
      configured: "configured",
      days: "days",
      milliseconds: "ms",
      trace: "Trace",
      status: "Status",
      severity: "Severity",
      category: "Category",
      source: "Source",
      summary: "Summary",
      action: "Action",
      event: "Event",
      path: "Path",
      priority: "Priority",
    },
    errors: {
      unknown_error: "Unknown error",
      network_error: "Network error",
      login_required: "Sign in before running this action.",
      invalid_credentials: "Invalid username or password.",
      runtime_report_unavailable: "Runtime diagnostics are temporarily unavailable.",
      request_logs_unavailable: "Request log lookup is temporarily unavailable.",
      users_unavailable: "User data is temporarily unavailable.",
      worker_unavailable: "Worker service is temporarily unavailable.",
      request_failed_503: "Runtime diagnostics are temporarily unavailable.",
      invalid_restore_drill_verification: "Restore drill verification artifact is invalid.",
      invalid_restore_drill_manifest: "Restore drill manifest artifact is invalid.",
      api_circuit_open: "API requests are temporarily paused while the gateway recovers.",
    },
    roles: {
      admin: "Admin",
      user: "Developer",
      developer: "Developer",
    },
    toolbar: {
      language: "Language",
      theme: "Theme",
      english: "EN",
      turkish: "TR",
      light: "Light",
      dark: "Dark",
    },
    pageTitles: {
      home: "AppFoundryLab | Home",
      test: "AppFoundryLab | Test Template",
    },
    hero: {
      eyebrow: "AppFoundryLab - Polyglot Full-Stack Boilerplate",
      title: "Start with a running system.",
      description:
        "AppFoundryLab brings Astro + Svelte frontend, Go services, a Rust worker, the data layer, observability, restore drill flows, and CI/CD into one repo with a realistic full-stack baseline.",
      supportingText:
        "The goal is not a showcase demo. The goal is to give a new contributor a clear starting point to boot the system locally, inspect the operational surface, and move changes forward safely.",
    },
    testPage: {
      title: "Test Template Page",
      description:
        "Smoke-test surface for the running stack and the sample restore drill artifacts.",
    },
    restoreDrill: {
      title: "Restore Drill Artifact Preview",
      description: "Static preview for regression coverage of bundle verification artifacts.",
      marker: "Marker",
      status: "Status",
      users: "Users",
      requestLogs: "Request Logs",
      fixtureManifest: "Fixture Manifest",
      loading: "Loading sample restore drill artifacts...",
    },
    systemStatus: {
      title: "System Status",
      loading: "Loading backend status...",
      gatewayState: "Gateway State",
      postgresRedis: "Postgres / Redis",
      grpcWorker: "gRPC Worker",
      authTitle: "JWT Login (RBAC)",
      authHelper:
        "Use admin or developer credentials from ./scripts/bootstrap.sh output or .env.docker.local.",
      username: "username",
      password: "password",
      loginButton: "Login + Load Users",
      authenticatedAsRole: "Authenticated as role",
      adminTitle: "Admin Runtime Diagnostics",
      diagnosticsPending: "Diagnostics will load after admin authentication.",
      exportSnapshot: "Export snapshot",
      downloadRuntimeReport: "Download JSON report",
      downloadIncidentReport: "Download incident report",
      copyIncidentSummary: "Copy incident summary",
      profile: "Profile",
      rateLimitStore: "Rate Limit Store",
      workerTls: "Worker TLS",
      legacyApi: "Legacy API",
      strictDependencies: "Strict dependencies",
      localAuthMode: "Local auth mode",
      signedLoggerIngest: "Signed logger ingest",
      loadShedding: "Load shedding",
      autoMigrate: "Auto migrate",
      redisFailureMode: "Redis failure mode",
      incidentSink: "Incident sink",
      incidentWebhook: "Incident webhook",
      incidentRetention: "Incident retention",
      diagnosticsCache: "Diagnostics cache",
      readyCache: "Ready cache",
      staleIfError: "Stale-if-error",
      reportVersion: "Report version",
      nextActions: "Next actions",
      runbookMapping: "Runbook mapping",
      operationalAlerts: "Operational Alerts",
      active: "Active",
      highest: "Highest",
      breaches: "Breaches",
      lastTriggered: "Last triggered",
      requests: "Requests",
      serverErrors: "5xx Errors",
      averageLatency: "Avg Latency",
      loadShed: "Load Shed",
      errorRate: "Error rate",
      inflightCurrent: "Inflight current",
      inflightPeak: "Inflight peak",
      latencySamples: "Latency samples",
      recentTrend: "Recent Trend",
      trendRequests: "Req",
      trendErrors: "Err",
      trendLatency: "Lat",
      healthCorrelation: "Health Correlation",
      checks: "Checks",
      httpStatus: "HTTP status",
      cache: "Cache",
      cacheAge: "Cache age",
      traceFlow: "Trace Flow",
      header: "Header",
      enabled: "Enabled",
      forwardedToLogger: "Forwarded to logger",
      loggerHeader: "Logger header",
      logField: "Log field",
      gatewayLoggerQueue: "Gateway Logger Queue",
      queue: "Queue",
      workers: "Workers",
      retryMax: "Retry max",
      dropped: "Dropped",
      traceLookup: "Trace Lookup",
      traceFilter: "Trace filter",
      latestTraceSummary: "Latest request logs from the logger backend",
      tracePlaceholder: "traceId",
      search: "Search",
      latest: "Latest",
      useTrace: "Use {traceId}",
      loadingRequestLogs: "Loading request logs...",
      noTraceMatch: "No request logs matched this trace.",
      noRequestLogs: "No request logs recorded yet.",
      loggerService: "Logger Service",
      configured: "Configured",
      reachable: "Reachable",
      health: "Health",
      processed: "Processed",
      retried: "Retried",
      inflightWorkers: "Inflight workers",
      dropRatio: "Drop ratio",
      loggerError: "Logger error",
      incidentJournal: "Incident Journal",
      sink: "Sink",
      totalEvents: "Total events",
      activeEvents: "Active events",
      lastEventStatus: "Last event status",
      dispatchFailures: "Dispatch failures",
      lastDispatch: "Last dispatch",
      latestEvent: "Latest event",
      incidentJournalError: "Incident journal error",
      recentIncidentEvents: "Recent Incident Events",
      noIncidentEvents: "No persistent incident events recorded yet.",
      lastSeen: "Last seen",
      useTraceButton: "Use trace",
      warnings: "Warnings",
      protectedUsers: "Protected Users Endpoint",
      noUsers: "No users loaded yet. Authenticate first.",
      fibonacciTitle: "gRPC Worker (Fibonacci)",
      compute: "Compute",
      result: "Result",
      clipboardIncident: "Incident",
      clipboardAlerts: "Alerts",
      clipboardRunbooks: "Runbooks",
    },
  },
  tr: {
    common: {
      error: "Hata",
      none: "yok",
      notAvailable: "yok",
      unknown: "Bilinmiyor",
      yes: "evet",
      no: "hayir",
      on: "acik",
      off: "kapali",
      enabled: "etkin",
      disabled: "devre disi",
      required: "zorunlu",
      optional: "opsiyonel",
      configured: "yapilandirildi",
      days: "gun",
      milliseconds: "ms",
      trace: "Iz",
      status: "Durum",
      severity: "Seviye",
      category: "Kategori",
      source: "Kaynak",
      summary: "Ozet",
      action: "Aksiyon",
      event: "Olay",
      path: "Yol",
      priority: "Oncelik",
    },
    errors: {
      unknown_error: "Bilinmeyen hata",
      network_error: "Ag hatasi",
      login_required: "Bu islemi calistirmadan once giris yapin.",
      invalid_credentials: "Kullanici adi veya sifre gecersiz.",
      runtime_report_unavailable: "Runtime diagnostics gecici olarak kullanilamiyor.",
      request_logs_unavailable: "Istek kaydi sorgusu gecici olarak kullanilamiyor.",
      users_unavailable: "Kullanici verisi gecici olarak kullanilamiyor.",
      worker_unavailable: "Worker servisi gecici olarak kullanilamiyor.",
      request_failed_503: "Runtime diagnostics gecici olarak kullanilamiyor.",
      invalid_restore_drill_verification: "Restore drill dogrulama dosyasi gecersiz.",
      invalid_restore_drill_manifest: "Restore drill manifest dosyasi gecersiz.",
      api_circuit_open: "Gateway toparlanirken API istekleri gecici olarak durduruldu.",
    },
    roles: {
      admin: "Yonetici",
      user: "Gelistirici",
      developer: "Gelistirici",
    },
    toolbar: {
      language: "Dil",
      theme: "Tema",
      english: "EN",
      turkish: "TR",
      light: "Acik",
      dark: "Koyu",
    },
    pageTitles: {
      home: "AppFoundryLab | Ana Sayfa",
      test: "AppFoundryLab | Test Sayfasi",
    },
    hero: {
      eyebrow: "AppFoundryLab - Polyglot Full-Stack Boilerplate",
      title: "Calisan bir sistemle baslayin.",
      description:
        "AppFoundryLab; Astro + Svelte frontend, Go servisleri, Rust worker, veri katmani, gozlemlenebilirlik, restore drill akislari ve CI/CD zincirini gercek hayata yakin tek bir full-stack tabanda toplar.",
      supportingText:
        "Amac vitrin demosu degil. Amac; yeni bir gelistiriciye sistemi yerelde hizli kaldirma, operasyon yuzeyini net gorme ve degisikligi guvenli sekilde ilerletme zemini vermektir.",
    },
    testPage: {
      title: "Test Sablon Sayfasi",
      description: "Calisan stack ve ornek restore drill artifact'lari icin smoke-test yuzeyi.",
    },
    restoreDrill: {
      title: "Restore Drill Artifact Onizleme",
      description: "Bundle dogrulama artifact'lari icin sabit regression onizlemesi.",
      marker: "Isaretleyici",
      status: "Durum",
      users: "Kullanicilar",
      requestLogs: "Istek Kayitlari",
      fixtureManifest: "Fixture Manifesti",
      loading: "Ornek restore drill artifact'lari yukleniyor...",
    },
    systemStatus: {
      title: "Sistem Durumu",
      loading: "Backend durumu yukleniyor...",
      gatewayState: "Gateway Durumu",
      postgresRedis: "Postgres / Redis",
      grpcWorker: "gRPC Worker",
      authTitle: "JWT Girisi (RBAC)",
      authHelper:
        "./scripts/bootstrap.sh cikisindaki veya .env.docker.local icindeki admin/gelistirici bilgilerini kullanin.",
      username: "kullanici",
      password: "sifre",
      loginButton: "Giris Yap ve Kullanicilari Yukle",
      authenticatedAsRole: "Dogrulanan rol",
      adminTitle: "Yonetici Runtime Teshisleri",
      diagnosticsPending: "Yonetici girisinden sonra teshis verileri yuklenecek.",
      exportSnapshot: "Disa aktarma anlik goruntusu",
      downloadRuntimeReport: "JSON raporunu indir",
      downloadIncidentReport: "Incident raporunu indir",
      copyIncidentSummary: "Incident ozetini kopyala",
      profile: "Profil",
      rateLimitStore: "Rate Limit Deposu",
      workerTls: "Worker TLS",
      legacyApi: "Legacy API",
      strictDependencies: "Kati bagimliliklar",
      localAuthMode: "Yerel auth modu",
      signedLoggerIngest: "Imzali logger ingest",
      loadShedding: "Yuk azaltma",
      autoMigrate: "Otomatik migrate",
      redisFailureMode: "Redis hata modu",
      incidentSink: "Incident sink",
      incidentWebhook: "Incident webhook",
      incidentRetention: "Incident saklama",
      diagnosticsCache: "Teshis cache'i",
      readyCache: "Ready cache'i",
      staleIfError: "Hata durumunda stale",
      reportVersion: "Rapor surumu",
      nextActions: "Siradaki aksiyonlar",
      runbookMapping: "Runbook eslesmesi",
      operationalAlerts: "Operasyonel Uyarilar",
      active: "Aktif",
      highest: "En yuksek",
      breaches: "Ihlal sayisi",
      lastTriggered: "Son tetiklenme",
      requests: "Istekler",
      serverErrors: "5xx Hatalari",
      averageLatency: "Ort. Gecikme",
      loadShed: "Load Shed",
      errorRate: "Hata orani",
      inflightCurrent: "Mevcut inflight",
      inflightPeak: "Zirve inflight",
      latencySamples: "Gecikme ornekleri",
      recentTrend: "Son Egilim",
      trendRequests: "Istk",
      trendErrors: "Hata",
      trendLatency: "Gec",
      healthCorrelation: "Saglik Iliskisi",
      checks: "Kontroller",
      httpStatus: "HTTP durumu",
      cache: "Cache",
      cacheAge: "Cache yasi",
      traceFlow: "Iz Akisi",
      header: "Baslik",
      enabled: "Etkin",
      forwardedToLogger: "Logger'a iletildi",
      loggerHeader: "Logger basligi",
      logField: "Log alani",
      gatewayLoggerQueue: "Gateway Logger Kuyrugu",
      queue: "Kuyruk",
      workers: "Worker'lar",
      retryMax: "Retry siniri",
      dropped: "Dusurulen",
      traceLookup: "Iz Aramasi",
      traceFilter: "Iz filtresi",
      latestTraceSummary: "Logger backend'indeki son istek kayitlari",
      tracePlaceholder: "traceId",
      search: "Ara",
      latest: "Son",
      useTrace: "{traceId} kullan",
      loadingRequestLogs: "Istek kayitlari yukleniyor...",
      noTraceMatch: "Bu izle eslesen istek kaydi yok.",
      noRequestLogs: "Henuz istek kaydi yok.",
      loggerService: "Logger Servisi",
      configured: "Yapilandirildi",
      reachable: "Erisilebilir",
      health: "Saglik",
      processed: "Islenen",
      retried: "Tekrar denenen",
      inflightWorkers: "Inflight worker",
      dropRatio: "Dusme orani",
      loggerError: "Logger hatasi",
      incidentJournal: "Incident Gunlugu",
      sink: "Sink",
      totalEvents: "Toplam olay",
      activeEvents: "Aktif olay",
      lastEventStatus: "Son olay durumu",
      dispatchFailures: "Dispatch hatalari",
      lastDispatch: "Son dispatch",
      latestEvent: "En son olay",
      incidentJournalError: "Incident gunlugu hatasi",
      recentIncidentEvents: "Son Incident Olaylari",
      noIncidentEvents: "Henuz kalici incident olayi kaydedilmedi.",
      lastSeen: "Son gorulme",
      useTraceButton: "Iz kullan",
      warnings: "Uyarilar",
      protectedUsers: "Korumali Kullanici Ucu",
      noUsers: "Henuz kullanici yuklenmedi. Once giris yapin.",
      fibonacciTitle: "gRPC Worker (Fibonacci)",
      compute: "Hesapla",
      result: "Sonuc",
      clipboardIncident: "Incident",
      clipboardAlerts: "Uyarilar",
      clipboardRunbooks: "Runbook'lar",
    },
  },
} as const;

export function getCopy(locale: Locale) {
  return copy[locale] ?? copy[DEFAULT_LOCALE];
}

export function getPageTitle(locale: Locale, key: PageTitleKey): string {
  return getCopy(locale).pageTitles[key];
}

export function formatDateTime(locale: Locale, value: string): string {
  if (!value) {
    return getCopy(locale).common.notAvailable;
  }

  return new Intl.DateTimeFormat(LOCALE_TAGS[locale], {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function formatTime(locale: Locale, value: string): string {
  return new Intl.DateTimeFormat(LOCALE_TAGS[locale], {
    timeStyle: "short",
  }).format(new Date(value));
}

export function formatPercent(locale: Locale, value: number, digits = 2): string {
  return new Intl.NumberFormat(LOCALE_TAGS[locale], {
    style: "percent",
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  }).format(value);
}

export function formatDecimal(locale: Locale, value: number, digits = 1): string {
  return new Intl.NumberFormat(LOCALE_TAGS[locale], {
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  }).format(value);
}

export function formatDurationMs(locale: Locale, value: number, digits = 0): string {
  return `${formatDecimal(locale, value, digits)} ${getCopy(locale).common.milliseconds}`;
}

export function booleanLabel(
  locale: Locale,
  value: boolean,
  mode: "yesNo" | "enabledDisabled" | "onOff" | "requiredOptional",
): string {
  const common = getCopy(locale).common;
  if (mode === "enabledDisabled") {
    return value ? common.enabled : common.disabled;
  }
  if (mode === "onOff") {
    return value ? common.on : common.off;
  }
  if (mode === "requiredOptional") {
    return value ? common.required : common.optional;
  }
  return value ? common.yes : common.no;
}

export function formatRole(locale: Locale, value: string): string {
  const roles = getCopy(locale).roles as Record<string, string>;
  return roles[value] ?? value;
}

export function translateError(locale: Locale, code: string): string {
  const errors = getCopy(locale).errors as Record<string, string>;
  return errors[code] ?? code;
}

export function renderTemplate(template: string, values: Record<string, string>): string {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => values[key] ?? "");
}


from django.contrib import admin
from django.urls import path, include
from django.conf import settings
from django.conf.urls.static import static
from django.urls import re_path
from rest_framework import permissions
from drf_yasg.views import get_schema_view
from drf_yasg import openapi
from django.http import JsonResponse


schema_view = get_schema_view(
   openapi.Info(
      title="article Api",
      default_version='v1',
      description="api for article",
      terms_of_service="https://www.google.com/policies/terms/",
      contact=openapi.Contact(email="6ix.mobin@gmail.com"),
      license=openapi.License(name="MIT License"),
   ),
   public=True,
   permission_classes=(permissions.AllowAny,),
)


def health_check(request):
    return JsonResponse({
        'status': 'ok',
        'service': 'article-service',
        'database': 'connected' if settings.DATABASES['default']['ENGINE'] != 'django.db.backends.sqlite3' else 'sqlite'
    })


urlpatterns = [
    path('health/', health_check, name='health'),
    path('admin/', admin.site.urls),
    path('article/', include('articles.urls')),
    # path('api-auth/', include('rest_framework.urls')),
    path('swagger/output.json', schema_view.without_ui(cache_timeout=0), name='schema-json'),
    path('swagger/', schema_view.with_ui('swagger', cache_timeout=0), name='schema-swagger-ui'),
    path('redoc/', schema_view.with_ui('redoc', cache_timeout=0), name='schema-redoc'),
]


if settings.DEBUG:
    urlpatterns += static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
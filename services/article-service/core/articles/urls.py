from django.urls import path, include
from . import views
from django.views.generic import TemplateView
from django.views.generic.base import RedirectView


app_name = 'article'

urlpatterns = [
   

    # baraye estefade az api dar appe blog yek url misazim ke maalom she az api estefade mikone in page
    # baad baiad dakhel app blog yek folder besazim be esme api va dakhl on views va url besazim va url paeein ro vasl konim be folder api
    # dakhel folder api ham view haye marbot be api ro misazim
    path('api/', include('articles.api.v1.urls')),


]
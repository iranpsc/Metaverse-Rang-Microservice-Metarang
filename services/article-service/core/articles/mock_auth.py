# articles/mock_auth.py
from django.contrib.auth.backends import BaseBackend
from django.contrib.auth.models import User

class MockJWTAuthenticationBackend(BaseBackend):
    def authenticate(self, request, token=None, **kwargs):
  
        user, _ = User.objects.get_or_create(username='testuser')
        request.remote_user_payload = {
            'username': 'testuser',
            'citizenId': 'hm-2000003',
            'avatar': 'https://example.com/avatar.jpg',
            'slug_level': 5,
            'image_level': 'https://example.com/level5.png',
        }
        return user

    def get_user(self, user_id):
        try:
            return User.objects.get(pk=user_id)
        except User.DoesNotExist:
            return None
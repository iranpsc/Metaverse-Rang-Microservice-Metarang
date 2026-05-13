

# prod
# kare in file ine ke token ro daryaft kone
from rest_framework.authentication import BaseAuthentication
from rest_framework.exceptions import AuthenticationFailed
from .services import RemoteJWTAuthenticationBackend

class RemoteJWTAuthentication(BaseAuthentication):
    def authenticate(self, request):
        auth_header = request.headers.get('Authorization', '')
        if not auth_header.startswith('Bearer '):
            return None
        
        token = auth_header.split(' ')[1]
        backend = RemoteJWTAuthenticationBackend()
        user = backend.authenticate(request, token=token)
        
        if user is None:
            raise AuthenticationFailed('Invalid token.')
        
        return (user, None)
    
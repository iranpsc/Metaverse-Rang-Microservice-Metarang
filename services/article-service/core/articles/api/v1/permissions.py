from rest_framework import permissions
from rest_framework.permissions import BasePermission



class IsOwnerOrReadOnly(BasePermission):
    
    def has_object_permission(self, request, view, obj):
        if request.method in permissions.SAFE_METHODS:
            return True
        
        user_payload = getattr(request, 'remote_user_payload', None)
        if not user_payload:
            return False
        
        citizen_id = user_payload.get('citizenId', '')        
        return obj.citizenId == citizen_id

    

class IsAuthenticatedRemote(BasePermission):
    
    def has_permission(self, request, view):
        if request.method in permissions.SAFE_METHODS:
            return True   
             
        user_payload = getattr(request, 'remote_user_payload', None)
        return user_payload is not None




class IsAdminCitizenOrReadOnly(permissions.BasePermission):

    def has_permission(self, request, view):
        if request.method in permissions.SAFE_METHODS:
            return True

        user_payload = getattr(request, 'remote_user_payload', {})
        if not user_payload:
            return False

        citizen_id = user_payload.get('citizenId', '')
        return citizen_id == "hm-2000003"
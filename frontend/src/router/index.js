import { createRouter, createWebHistory } from 'vue-router'
import HomepageIndex from "@/views/homepage/HomepageIndex.vue";
import CreateIcon from "@/components/navbar/icons/CreateIcon.vue";
import CreateIndex from "@/views/create/CreateIndex.vue";
import FriendIndex from "@/views/friend/FriendIndex.vue";
import LoginIndex from "@/views/user/account/LoginIndex.vue";
import RegisterIndex from "@/views/user/account/RegisterIndex.vue";
import ProfileIndex from "@/views/user/profile/ProfileIndex.vue";
import NotFoundIndex from "@/views/error/NotFoundIndex.vue";
import SpaceIndex from "@/views/user/space/SpaceIndex.vue";
import {useUserStore} from "@/stores/user.js";
import UpdateCharacter from "@/views/create/character/UpdateCharacter.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: HomepageIndex,
      name: 'homepage-index',
      meta: {
        needLogin: false
      }
    },
    {
      path: '/create/',
      component: CreateIndex,
      name: 'create-index',
      meta: {
        needLogin: true
      }
    },
    {
      path: '/create/character/update/:character_id',
      component: UpdateCharacter,
      name: 'update-character-index',
      meta: {
        needLogin: true
      }
    },
    {
      path: '/friend/',
      component: FriendIndex,
      name: 'friend-index',
      meta: {
        needLogin: true
      }
    },
    {
      path: '/user/account/login/',
      component: LoginIndex,
      name: 'user-account-login-index',
      meta: {
        needLogin: false
      }
    },
    {
      path: '/user/account/register/',
      component: RegisterIndex,
      name: 'user-account-register-index',
      meta: {
        needLogin: false
      }
    },
    {
      path: '/user/space/:user_id/',
      component: SpaceIndex,
      name: 'user-space-index',
      meta: {
        needLogin: true
      }
    },
    {
      path: '/user/profile/',
      component: ProfileIndex,
      name: 'user-profile-index',
      meta: {
        needLogin: true
      }
    },
    {
      path: '/404/',
      component: NotFoundIndex,
      name: '404',
      meta: {
        needLogin: false
      }
    },
    {
      path: '/:pathMatch(.*)*',
      component: NotFoundIndex,
      name: 'not-found',
      meta: {
        needLogin: false
      }
    },
  ],
})

router.beforeEach((to) => {
  const user = useUserStore()
  if(to.meta.needLogin && user.hasPulledUserInfo && !user.isLogin()) {
    return {
      name: 'user-account-login-index',
    }
  }
  return true
})

export default router

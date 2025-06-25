<script lang="ts" setup>
import { onMounted } from 'vue';
import { lunaCommunicator } from '@/utils/lunaBus.ts';
import { LUNA_MESSAGE_TYPE } from '@/types/postmessage.type';


onMounted(() => {
    const handLunaCommand = (message: any) => {
        console.log('Received Luna command:', message);

    }
    lunaCommunicator.onLuna(LUNA_MESSAGE_TYPE.CMD, handLunaCommand);
});

</script>


<template>
    <div>
        connect view
    </div>

</template>